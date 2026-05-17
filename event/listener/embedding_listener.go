package listener

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/evo-lee/evo-sonic/consts"
	"github.com/evo-lee/evo-sonic/dal/vector"
	"github.com/evo-lee/evo-sonic/event"
	"github.com/evo-lee/evo-sonic/log"
	"github.com/evo-lee/evo-sonic/model/property"
	"github.com/evo-lee/evo-sonic/service"
	ai "github.com/evo-lee/evo-sonic/service/ai"
)

type EmbeddingListener struct {
	postService       service.PostService
	embeddingProvider ai.EmbeddingProvider
	vectorStore       ai.VectorStore
	optionService     service.OptionService
}

// NewEmbeddingListener subscribes to PostUpdateEvent and asynchronously
// embeds published posts into the vector store.
func NewEmbeddingListener(
	bus event.Bus,
	db *gorm.DB,
	postService service.PostService,
	embeddingProvider ai.EmbeddingProvider,
	optionService service.OptionService,
) {
	l := &EmbeddingListener{
		postService:       postService,
		embeddingProvider: embeddingProvider,
		vectorStore:       vector.NewDBStore(db),
		optionService:     optionService,
	}
	bus.Subscribe(event.PostUpdateEventName, l.handle)
}

func (l *EmbeddingListener) handle(ctx context.Context, e event.Event) error {
	postID := e.(*event.PostUpdateEvent).PostID

	post, err := l.postService.GetByPostID(ctx, postID)
	if err != nil {
		return err
	}

	// Remove vector when post is recycled/deleted.
	if post.Status == consts.PostStatusRecycle {
		go func() {
			if delErr := l.vectorStore.Delete(context.Background(), postID); delErr != nil {
				log.Error("embedding delete failed", zap.Int32("postID", postID), zap.Error(delErr))
			}
		}()
		return nil
	}

	// Only embed published / intimate posts with content.
	if post.Status != consts.PostStatusPublished && post.Status != consts.PostStatusIntimate {
		return nil
	}
	if post.OriginalContent == "" {
		return nil
	}

	// Embed asynchronously so the event bus is not blocked.
	go func() {
		bgCtx := context.Background()
		model := l.optionService.GetOrByDefault(bgCtx, property.AIEmbeddingModel).(string)
		if model == "" {
			model = "text-embedding-3-small"
		}

		text := post.Title + "\n\n" + post.OriginalContent
		vec, err := l.embeddingProvider.Embed(bgCtx, text)
		if err != nil {
			log.Error("embedding failed", zap.Int32("postID", postID), zap.Error(err))
			return
		}
		if err := l.vectorStore.Upsert(bgCtx, postID, model, vec); err != nil {
			log.Error("embedding upsert failed", zap.Int32("postID", postID), zap.Error(err))
		}
	}()

	return nil
}
