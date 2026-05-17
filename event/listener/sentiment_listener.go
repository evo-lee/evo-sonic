package listener

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/evo-lee/evo-sonic/consts"
	"github.com/evo-lee/evo-sonic/dal"
	"github.com/evo-lee/evo-sonic/event"
	"github.com/evo-lee/evo-sonic/log"
	ai "github.com/evo-lee/evo-sonic/service/ai"
)

const sentimentSystemPrompt = `You are a content moderation classifier.
Classify the comment as one of: positive, neutral, negative, spam.
Reply with exactly one word (no punctuation, lowercase).`

// SentimentListener listens for new comments and runs AI sentiment analysis.
// Comments classified as negative or spam are held for manual review (Auditing status).
type SentimentListener struct {
	provider ai.Provider
}

// NewSentimentListener subscribes to CommentNewEvent via an async worker pool
// so AI latency never blocks comment submission.
func NewSentimentListener(bus event.Bus, provider ai.Provider) {
	l := &SentimentListener{provider: provider}
	pool := event.NewAsyncWorkerPool(bus, 2, 64)
	pool.SubscribeAsync(event.CommentNewEventName, l.handle)
}

func (l *SentimentListener) handle(ctx context.Context, e event.Event) error {
	comment := e.(*event.CommentNewEvent).Comment

	resp, err := l.provider.Complete(context.Background(), ai.CompletionRequest{
		System:    sentimentSystemPrompt,
		Prompt:    comment.Content,
		MaxTokens: 10,
	})
	if err != nil {
		// Non-fatal: AI unavailable — leave comment status as-is.
		log.Error("sentiment analysis failed", zap.Int32("commentID", comment.ID), zap.Error(err))
		return nil
	}

	label := strings.TrimSpace(strings.ToLower(resp.Content))
	if label != "negative" && label != "spam" {
		return nil
	}

	// Hold comment for manual review.
	bgCtx := context.Background()
	q := dal.GetQueryByCtx(bgCtx)
	_, err = q.Comment.WithContext(bgCtx).
		Where(q.Comment.ID.Eq(comment.ID)).
		Update(q.Comment.Status, consts.CommentStatusAuditing)
	if err != nil {
		log.Error("sentiment update status failed", zap.Int32("commentID", comment.ID), zap.Error(err))
	}
	return nil
}
