package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	hertzapp "github.com/cloudwego/hertz/pkg/app"

	"github.com/go-sonic/sonic/handler/web"
	"github.com/go-sonic/sonic/model/property"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/service/ai"
	"github.com/go-sonic/sonic/util/xerr"
)

type AIHandler struct {
	contentService ai.ContentService
	optionService  service.OptionService
}

func NewAIHandler(contentService ai.ContentService, optionService service.OptionService) *AIHandler {
	return &AIHandler{contentService: contentService, optionService: optionService}
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
