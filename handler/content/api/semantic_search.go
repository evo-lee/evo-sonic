package api

import (
	"gorm.io/gorm"

	"github.com/go-sonic/sonic/dal/vector"
	"github.com/go-sonic/sonic/handler/web"
	"github.com/go-sonic/sonic/service"
	ai "github.com/go-sonic/sonic/service/ai"
	"github.com/go-sonic/sonic/util/xerr"
)

type SemanticSearchHandler struct {
	embeddingProvider ai.EmbeddingProvider
	vectorStore       ai.VectorStore
	postService       service.PostService
}

func NewSemanticSearchHandler(
	db *gorm.DB,
	embeddingProvider ai.EmbeddingProvider,
	postService service.PostService,
) *SemanticSearchHandler {
	return &SemanticSearchHandler{
		embeddingProvider: embeddingProvider,
		vectorStore:       vector.NewDBStore(db),
		postService:       postService,
	}
}

// Search handles GET /api/content/search/semantic?q=<query>&limit=10
func (h *SemanticSearchHandler) Search(ctx web.Context) (interface{}, error) {
	query, _ := ctx.Query("q")
	if query == "" {
		return nil, xerr.BadParam.New("q is required")
	}
	limit := 10
	if l, ok := ctx.Query("limit"); ok && l != "" {
		if n, err := parseInt(l); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}

	reqCtx := ctx.RequestContext()
	vec, err := h.embeddingProvider.Embed(reqCtx, query)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusInternalServerError).WithMsg("embedding failed")
	}

	results, err := h.vectorStore.Search(reqCtx, vec, limit)
	if err != nil {
		return nil, err
	}

	postIDs := make([]int32, len(results))
	for i, r := range results {
		postIDs[i] = r.PostID
	}
	posts, err := h.postService.GetByPostIDs(reqCtx, postIDs)
	if err != nil {
		return nil, err
	}

	type item struct {
		PostID int32   `json:"post_id"`
		Title  string  `json:"title"`
		Slug   string  `json:"slug"`
		Score  float64 `json:"score"`
	}
	items := make([]item, 0, len(results))
	for _, r := range results {
		if post, ok := posts[r.PostID]; ok {
			items = append(items, item{
				PostID: r.PostID,
				Title:  post.Title,
				Slug:   post.Slug,
				Score:  r.Score,
			})
		}
	}
	return map[string]interface{}{"posts": items}, nil
}

func parseInt(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, xerr.BadParam.New("not a number")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
