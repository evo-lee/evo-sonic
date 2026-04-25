package listener

import (
	"context"

	"go.uber.org/zap"

	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/dal"
	"github.com/go-sonic/sonic/event"
	"github.com/go-sonic/sonic/log"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/service/ai"
)

type AISummaryListener struct {
	postService    service.PostService
	contentService ai.ContentService
}

// NewAISummaryListener subscribes to PostUpdateEvent and backfills the summary
// field using AI when the post is published and has no existing summary.
func NewAISummaryListener(bus event.Bus, postService service.PostService, contentService ai.ContentService) {
	l := &AISummaryListener{
		postService:    postService,
		contentService: contentService,
	}
	bus.Subscribe(event.PostUpdateEventName, l.handle)
}

func (l *AISummaryListener) handle(ctx context.Context, e event.Event) error {
	postID := e.(*event.PostUpdateEvent).PostID

	post, err := l.postService.GetByPostID(ctx, postID)
	if err != nil {
		return err
	}
	// Only fill summary for published posts that have no existing summary.
	if post.Status != consts.PostStatusPublished {
		return nil
	}
	if post.Summary != "" {
		return nil
	}
	if post.OriginalContent == "" {
		return nil
	}

	// Run in a detached goroutine so AI latency does not block the event bus.
	go func() {
		bgCtx := context.Background()
		summary, err := l.contentService.Summarize(bgCtx, post.OriginalContent)
		if err != nil {
			log.Error("ai summary failed", zap.Int32("postID", postID), zap.Error(err))
			return
		}
		q := dal.GetQueryByCtx(bgCtx)
		_, err = q.Post.WithContext(bgCtx).Where(q.Post.ID.Eq(postID)).Update(q.Post.Summary, summary)
		if err != nil {
			log.Error("ai summary update failed", zap.Int32("postID", postID), zap.Error(err))
		}
	}()

	return nil
}
