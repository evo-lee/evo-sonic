package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	hertzapp "github.com/cloudwego/hertz/pkg/app"
	"gorm.io/gorm"

	"github.com/evo-lee/evo-sonic/dal"
	"github.com/evo-lee/evo-sonic/dal/vector"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/property"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/service/ai"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type AIHandler struct {
	contentService    ai.ContentService
	optionService     service.OptionService
	embeddingProvider ai.EmbeddingProvider
	vectorStore       ai.VectorStore
	postService       service.PostService
	workflowService   ai.WorkflowService
}

func NewAIHandler(
	db *gorm.DB,
	contentService ai.ContentService,
	optionService service.OptionService,
	embeddingProvider ai.EmbeddingProvider,
	postService service.PostService,
	workflowService ai.WorkflowService,
) *AIHandler {
	return &AIHandler{
		contentService:    contentService,
		optionService:     optionService,
		embeddingProvider: embeddingProvider,
		vectorStore:       vector.NewDBStore(db),
		postService:       postService,
		workflowService:   workflowService,
	}
}

// ── Non-streaming endpoints ────────────────────────────────────────────────

type summarizeRequest struct {
	Content string `json:"content"`
}

type suggestTagsRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type polishRequest struct {
	Content string `json:"content"`
}

func (h *AIHandler) Summarize(ctx web.Context) (interface{}, error) {
	var req summarizeRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	if req.Content == "" {
		return nil, xerr.BadParam.New("content is required")
	}
	summary, err := h.contentService.Summarize(ctx.RequestContext(), req.Content)
	if err != nil {
		return nil, err
	}
	return map[string]string{"summary": summary}, nil
}

func (h *AIHandler) SuggestTags(ctx web.Context) (interface{}, error) {
	var req suggestTagsRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	if req.Content == "" {
		return nil, xerr.BadParam.New("content is required")
	}
	tags, err := h.contentService.SuggestTags(ctx.RequestContext(), req.Title, req.Content)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"tags": tags}, nil
}

func (h *AIHandler) Polish(ctx web.Context) (interface{}, error) {
	var req polishRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	if req.Content == "" {
		return nil, xerr.BadParam.New("content is required")
	}
	polished, err := h.contentService.Polish(ctx.RequestContext(), req.Content)
	if err != nil {
		return nil, err
	}
	return map[string]string{"content": polished}, nil
}

// ── Config endpoints ───────────────────────────────────────────────────────

type aiConfigResponse struct {
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"` // masked
	Model    string `json:"model"`
	BaseURL  string `json:"base_url"`
}

type aiConfigRequest struct {
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
	Model    string `json:"model"`
	BaseURL  string `json:"base_url"`
}

func (h *AIHandler) GetConfig(ctx web.Context) (interface{}, error) {
	reqCtx := ctx.RequestContext()
	provider := h.optionService.GetOrByDefault(reqCtx, property.AIProvider).(string)
	apiKey := h.optionService.GetOrByDefault(reqCtx, property.AIAPIKey).(string)
	model := h.optionService.GetOrByDefault(reqCtx, property.AIModel).(string)
	baseURL := h.optionService.GetOrByDefault(reqCtx, property.AIBaseURL).(string)

	return &aiConfigResponse{
		Provider: provider,
		APIKey:   maskAPIKey(apiKey),
		Model:    model,
		BaseURL:  baseURL,
	}, nil
}

func (h *AIHandler) SaveConfig(ctx web.Context) (interface{}, error) {
	var req aiConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}

	optionMap := map[string]string{
		property.AIProvider.KeyValue: req.Provider,
		property.AIModel.KeyValue:    req.Model,
		property.AIBaseURL.KeyValue:  req.BaseURL,
	}
	// Only update api_key when explicitly provided (non-empty).
	if req.APIKey != "" {
		optionMap[property.AIAPIKey.KeyValue] = req.APIKey
	}

	if err := h.optionService.Save(ctx.RequestContext(), optionMap); err != nil {
		return nil, err
	}
	return map[string]string{"message": "AI config saved"}, nil
}

func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}

// ── Streaming (SSE) endpoints ──────────────────────────────────────────────

// SummarizeStream streams the summary as Server-Sent Events.
// This is a raw web.HandlerFunc (not wrapped) to allow direct response writing.
func (h *AIHandler) SummarizeStream(ctx web.Context) {
	var req summarizeRequest
	if err := ctx.BindJSON(&req); err != nil || req.Content == "" {
		writeSSEInputError(ctx, "content is required")
		return
	}
	ch, err := h.contentService.SummarizeStream(ctx.RequestContext(), req.Content)
	if err != nil {
		writeSSEInputError(ctx, err.Error())
		return
	}
	writeSSE(ctx, ch)
}

// PolishStream streams the polished content as Server-Sent Events.
func (h *AIHandler) PolishStream(ctx web.Context) {
	var req polishRequest
	if err := ctx.BindJSON(&req); err != nil || req.Content == "" {
		writeSSEInputError(ctx, "content is required")
		return
	}
	ch, err := h.contentService.PolishStream(ctx.RequestContext(), req.Content)
	if err != nil {
		writeSSEInputError(ctx, err.Error())
		return
	}
	writeSSE(ctx, ch)
}

// SuggestTagsStream streams tag suggestions as Server-Sent Events.
// Each chunk contains a fragment of the comma-separated tag list.
func (h *AIHandler) SuggestTagsStream(ctx web.Context) {
	var req suggestTagsRequest
	if err := ctx.BindJSON(&req); err != nil || req.Content == "" {
		writeSSEInputError(ctx, "content is required")
		return
	}
	ch, err := h.contentService.SuggestTagsStream(ctx.RequestContext(), req.Title, req.Content)
	if err != nil {
		writeSSEInputError(ctx, err.Error())
		return
	}
	writeSSE(ctx, ch)
}

// writeSSEInputError sends a single SSE error event and closes the stream.
// Used for pre-stream validation failures so clients handle errors uniformly.
func writeSSEInputError(ctx web.Context, msg string) {
	ctx.SetHeader("Content-Type", "text/event-stream")
	ctx.SetHeader("Cache-Control", "no-cache")
	ctx.SetHeader("X-Accel-Buffering", "no")
	data, _ := json.Marshal(map[string]string{"error": msg})
	fmt.Fprintf(ctx.Writer(), "event: error\ndata: %s\n\n", data)
}

// writeSSE writes StreamChunk events to the response as SSE.
// Each chunk becomes: data: {"text":"..."}\n\n
// Final message: data: [DONE]\n\n
func writeSSE(ctx web.Context, ch <-chan ai.StreamChunk) {
	ctx.SetHeader("Content-Type", "text/event-stream")
	ctx.SetHeader("Cache-Control", "no-cache")
	ctx.SetHeader("X-Accel-Buffering", "no")

	native := ctx.Native()
	hertzCtx, ok := native.(*hertzapp.RequestContext)
	if !ok {
		go func() { for range ch {} }() // drain so the provider goroutine can exit
		ctx.JSON(500, map[string]string{"error": "streaming not supported"})
		return
	}

	pr, pw := io.Pipe()
	hertzCtx.Response.SetBodyStream(pr, -1)

	go func() {
		defer pw.Close()
		for chunk := range ch {
			if chunk.Err != nil {
				data, _ := json.Marshal(map[string]string{"error": chunk.Err.Error()})
				fmt.Fprintf(pw, "event: error\ndata: %s\n\n", data)
				return
			}
			data, _ := json.Marshal(map[string]string{"text": chunk.Text})
			fmt.Fprintf(pw, "data: %s\n\n", data)
		}
		fmt.Fprintf(pw, "data: [DONE]\n\n")
	}()
}

// RelatedPosts returns the topK posts most similar to the given post, based on embeddings.
// GET /api/admin/ai/related-posts?postId=<id>&limit=5
func (h *AIHandler) RelatedPosts(ctx web.Context) (interface{}, error) {
	rawID, _ := ctx.Query("postId")
	if rawID == "" {
		return nil, xerr.BadParam.New("postId is required")
	}
	rawIDParsed, err := parseIntParam(rawID)
	if err != nil || rawIDParsed == 0 {
		return nil, xerr.BadParam.New("postId must be a positive integer")
	}
	postID := int32(rawIDParsed)
	limit := 5
	if l, ok := ctx.Query("limit"); ok && l != "" {
		if n, err := parseIntParam(l); err == nil && n > 0 && n <= 20 {
			limit = n
		}
	}

	reqCtx := ctx.RequestContext()
	post, err := h.postService.GetByPostID(reqCtx, postID)
	if err != nil {
		return nil, err
	}
	text := post.Title + "\n\n" + post.OriginalContent
	vec, err := h.embeddingProvider.Embed(reqCtx, text)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusInternalServerError).WithMsg("embedding failed")
	}

	// +1 so we can drop the post itself from results
	results, err := h.vectorStore.Search(reqCtx, vec, limit+1)
	if err != nil {
		return nil, err
	}

	postIDs := make([]int32, 0, len(results))
	for _, r := range results {
		if r.PostID != postID {
			postIDs = append(postIDs, r.PostID)
		}
	}
	if len(postIDs) > limit {
		postIDs = postIDs[:limit]
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
	scoreMap := make(map[int32]float64, len(results))
	for _, r := range results {
		scoreMap[r.PostID] = r.Score
	}
	items := make([]item, 0, len(postIDs))
	for _, id := range postIDs {
		if p, ok := posts[id]; ok {
			items = append(items, item{PostID: id, Title: p.Title, Slug: p.Slug, Score: scoreMap[id]})
		}
	}
	return map[string]interface{}{"posts": items}, nil
}

func parseIntParam(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, xerr.BadParam.New("not a number")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// SEOCheck validates a post and returns actionable SEO suggestions.
// POST /api/admin/ai/seo-check
func (h *AIHandler) SEOCheck(ctx web.Context) (interface{}, error) {
	var req struct {
		PostID int32 `json:"postId"`
	}
	if err := ctx.BindJSON(&req); err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	if req.PostID == 0 {
		return nil, xerr.BadParam.New("postId is required")
	}

	reqCtx := ctx.RequestContext()
	post, err := h.postService.GetByPostID(reqCtx, req.PostID)
	if err != nil {
		return nil, err
	}

	// Count tags via DAL
	q := dal.GetQueryByCtx(reqCtx)
	tagCount, _ := q.PostTag.WithContext(reqCtx).Where(q.PostTag.PostID.Eq(req.PostID)).Count()

	report, err := h.workflowService.CheckSEO(reqCtx, ai.SEOCheckRequest{
		PostID:          post.ID,
		Title:           post.Title,
		Summary:         post.Summary,
		OriginalContent: post.OriginalContent,
		MetaDescription: post.MetaDescription,
		TagCount:        int(tagCount),
	})
	if err != nil {
		return nil, err
	}
	return report, nil
}
