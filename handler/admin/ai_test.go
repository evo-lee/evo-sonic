package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/model/dto"
	"github.com/go-sonic/sonic/model/param"
	"github.com/go-sonic/sonic/model/property"
	"github.com/go-sonic/sonic/service/ai"
)

// ── mock ContentService ───────────────────────────────────────────────────────

type mockContentService struct {
	summarizeResult   string
	summarizeErr      error
	suggestTagsResult []string
	suggestTagsErr    error
	polishResult      string
	polishErr         error
}

func (m *mockContentService) Summarize(_ context.Context, _ string) (string, error) {
	return m.summarizeResult, m.summarizeErr
}
func (m *mockContentService) SuggestTags(_ context.Context, _, _ string) ([]string, error) {
	return m.suggestTagsResult, m.suggestTagsErr
}
func (m *mockContentService) Polish(_ context.Context, _ string) (string, error) {
	return m.polishResult, m.polishErr
}
func (m *mockContentService) SummarizeStream(_ context.Context, _ string) (<-chan ai.StreamChunk, error) {
	return nil, nil
}
func (m *mockContentService) PolishStream(_ context.Context, _ string) (<-chan ai.StreamChunk, error) {
	return nil, nil
}
func (m *mockContentService) SuggestTagsStream(_ context.Context, _, _ string) (<-chan ai.StreamChunk, error) {
	return nil, nil
}

// ── mock OptionService ────────────────────────────────────────────────────────

type mockOptionService struct {
	values map[string]string
}

func (m *mockOptionService) GetOrByDefault(_ context.Context, p property.Property) interface{} {
	if v, ok := m.values[p.KeyValue]; ok {
		return v
	}
	return p.DefaultValue
}
func (m *mockOptionService) GetOrByDefaultWithErr(_ context.Context, p property.Property, _ interface{}) (interface{}, error) {
	return m.GetOrByDefault(context.Background(), p), nil
}
func (m *mockOptionService) Save(_ context.Context, opts map[string]string) error {
	for k, v := range opts {
		m.values[k] = v
	}
	return nil
}
func (m *mockOptionService) GetBlogBaseURL(_ context.Context) (string, error)                        { return "", nil }
func (m *mockOptionService) IsEnabledAbsolutePath(_ context.Context) (bool, error)                   { return false, nil }
func (m *mockOptionService) GetPathSuffix(_ context.Context) (string, error)                         { return "", nil }
func (m *mockOptionService) GetArchivePrefix(_ context.Context) (string, error)                      { return "", nil }
func (m *mockOptionService) ListAllOption(_ context.Context) ([]*dto.Option, error)                  { return nil, nil }
func (m *mockOptionService) GetLinksPrefix(_ context.Context) (string, error)                        { return "", nil }
func (m *mockOptionService) GetPhotoPrefix(_ context.Context) (string, error)                        { return "", nil }
func (m *mockOptionService) GetJournalPrefix(_ context.Context) (string, error)                      { return "", nil }
func (m *mockOptionService) GetActivatedThemeID(_ context.Context) (string, error)                   { return "", nil }
func (m *mockOptionService) GetPostPermalinkType(_ context.Context) (consts.PostPermalinkType, error) { return "", nil }
func (m *mockOptionService) GetSheetPermalinkType(_ context.Context) (consts.SheetPermaLinkType, error) { return "", nil }
func (m *mockOptionService) GetIndexPageSize(_ context.Context) int                                  { return 0 }
func (m *mockOptionService) GetPostSort(_ context.Context) param.Sort                                { return param.Sort{} }
func (m *mockOptionService) GetPostSummaryLength(_ context.Context) int                              { return 0 }
func (m *mockOptionService) GetCategoryPrefix(_ context.Context) (string, error)                     { return "", nil }
func (m *mockOptionService) GetTagPrefix(_ context.Context) (string, error)                          { return "", nil }
func (m *mockOptionService) GetLinkPrefix(_ context.Context) (string, error)                         { return "", nil }
func (m *mockOptionService) GetSheetPrefix(_ context.Context) (string, error)                        { return "", nil }
func (m *mockOptionService) GetAttachmentType(_ context.Context) consts.AttachmentType               { return 0 }
func (m *mockOptionService) GetAdminURLPath(_ context.Context) (string, error)                       { return "", nil }

// ── mock web.Context ──────────────────────────────────────────────────────────

type mockWebCtx struct {
	body     []byte
	jsonCode int
	jsonBody interface{}
	headers  map[string]string
	writeBuf *bytes.Buffer
}

func newMockCtx(body string) *mockWebCtx {
	return &mockWebCtx{
		body:     []byte(body),
		headers:  make(map[string]string),
		writeBuf: &bytes.Buffer{},
	}
}

func (c *mockWebCtx) BindJSON(v any) error {
	return json.NewDecoder(bytes.NewReader(c.body)).Decode(v)
}
func (c *mockWebCtx) JSON(code int, v any)              { c.jsonCode = code; c.jsonBody = v }
func (c *mockWebCtx) SetHeader(k, v string)             { c.headers[k] = v }
func (c *mockWebCtx) RequestContext() context.Context   { return context.Background() }
func (c *mockWebCtx) Writer() io.Writer                 { return c.writeBuf }
func (c *mockWebCtx) AbortWithStatusJSON(code int, v any) { c.jsonCode = code; c.jsonBody = v }

// context.Context methods
func (c *mockWebCtx) Deadline() (time.Time, bool) { return time.Time{}, false }
func (c *mockWebCtx) Done() <-chan struct{}        { return nil }
func (c *mockWebCtx) Err() error                  { return nil }
func (c *mockWebCtx) Value(_ any) any             { return nil }

// Remaining web.Context methods — not used in handler unit tests.
func (c *mockWebCtx) Request() *http.Request                          { return nil }
func (c *mockWebCtx) Method() string                                  { return "POST" }
func (c *mockWebCtx) Path() string                                    { return "/" }
func (c *mockWebCtx) RawQuery() string                                { return "" }
func (c *mockWebCtx) ClientIP() string                                { return "127.0.0.1" }
func (c *mockWebCtx) Header(_ string) string                          { return "" }
func (c *mockWebCtx) ResponseHeader(_ string) string                  { return "" }
func (c *mockWebCtx) StatusCode() int                                 { return c.jsonCode }
func (c *mockWebCtx) Query(_ string) (string, bool)                   { return "", false }
func (c *mockWebCtx) Param(_ string) string                           { return "" }
func (c *mockWebCtx) Cookie(_ string) (string, error)                 { return "", nil }
func (c *mockWebCtx) SetCookie(_, _ string, _ int, _, _ string, _, _ bool) {}
func (c *mockWebCtx) MultipartForm() (*multipart.Form, error)         { return nil, nil }
func (c *mockWebCtx) Set(_ string, _ any)                             {}
func (c *mockWebCtx) Get(_ any) (any, bool)                           { return nil, false }
func (c *mockWebCtx) Bind(_ any) error                                { return nil }
func (c *mockWebCtx) BindQuery(_ any) error                           { return nil }
func (c *mockWebCtx) BindWith(_ any, _ any) error                     { return nil }
func (c *mockWebCtx) String(_ int, _ string)                          {}
func (c *mockWebCtx) Status(_ int)                                    {}
func (c *mockWebCtx) Redirect(_ int, _ string)                        {}
func (c *mockWebCtx) File(_ string)                                   {}
func (c *mockWebCtx) FormFile(_ string) (*multipart.FileHeader, error) { return nil, nil }
func (c *mockWebCtx) Abort()                                          {}
func (c *mockWebCtx) Next()                                           {}
func (c *mockWebCtx) Native() any                                     { return nil }

// ── mock EmbeddingProvider ────────────────────────────────────────────────────

type mockEmbeddingProvider struct{}

func (mockEmbeddingProvider) Embed(_ context.Context, _ string) ([]float32, error) {
	return []float32{0.1, 0.2, 0.3}, nil
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newHandler(cs *mockContentService, opts *mockOptionService) *AIHandler {
	// Pass nil db/workflowService — tests don't call RelatedPosts or SEOCheck.
	return NewAIHandler(nil, cs, opts, mockEmbeddingProvider{}, nil, nil)
}

// ── Summarize handler ─────────────────────────────────────────────────────────

func TestSummarizeHandler_EmptyContent_ReturnsBadParam(t *testing.T) {
	h := newHandler(&mockContentService{}, &mockOptionService{values: map[string]string{}})
	ctx := newMockCtx(`{"content":""}`)

	_, err := h.Summarize(ctx)
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestSummarizeHandler_Success_ReturnsSummary(t *testing.T) {
	cs := &mockContentService{summarizeResult: "short summary"}
	h := newHandler(cs, &mockOptionService{values: map[string]string{}})
	ctx := newMockCtx(`{"content":"long article text"}`)

	resp, err := h.Summarize(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := resp.(map[string]string)
	if !ok {
		t.Fatalf("expected map[string]string response, got %T", resp)
	}
	if m["summary"] != "short summary" {
		t.Errorf("expected 'short summary', got %q", m["summary"])
	}
}

func TestSummarizeHandler_ServiceError_Propagated(t *testing.T) {
	cs := &mockContentService{summarizeErr: errors.New("AI offline")}
	h := newHandler(cs, &mockOptionService{values: map[string]string{}})
	ctx := newMockCtx(`{"content":"text"}`)

	_, err := h.Summarize(ctx)
	if err == nil || !strings.Contains(err.Error(), "AI offline") {
		t.Errorf("expected 'AI offline' error, got %v", err)
	}
}

// ── SuggestTags handler ───────────────────────────────────────────────────────

func TestSuggestTagsHandler_EmptyContent_ReturnsBadParam(t *testing.T) {
	h := newHandler(&mockContentService{}, &mockOptionService{values: map[string]string{}})
	ctx := newMockCtx(`{"title":"title","content":""}`)

	_, err := h.SuggestTags(ctx)
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestSuggestTagsHandler_Success_ReturnsTags(t *testing.T) {
	cs := &mockContentService{suggestTagsResult: []string{"go", "testing"}}
	h := newHandler(cs, &mockOptionService{values: map[string]string{}})
	ctx := newMockCtx(`{"title":"Hello","content":"article body"}`)

	resp, err := h.SuggestTags(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := resp.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map response, got %T", resp)
	}
	tags, ok := m["tags"].([]string)
	if !ok || len(tags) != 2 {
		t.Errorf("expected 2 tags, got %v", m["tags"])
	}
}

// ── Polish handler ────────────────────────────────────────────────────────────

func TestPolishHandler_EmptyContent_ReturnsBadParam(t *testing.T) {
	h := newHandler(&mockContentService{}, &mockOptionService{values: map[string]string{}})
	ctx := newMockCtx(`{"content":""}`)

	_, err := h.Polish(ctx)
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestPolishHandler_Success_ReturnsPolishedContent(t *testing.T) {
	cs := &mockContentService{polishResult: "polished text"}
	h := newHandler(cs, &mockOptionService{values: map[string]string{}})
	ctx := newMockCtx(`{"content":"draft text"}`)

	resp, err := h.Polish(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := resp.(map[string]string)
	if !ok {
		t.Fatalf("expected map[string]string, got %T", resp)
	}
	if m["content"] != "polished text" {
		t.Errorf("expected 'polished text', got %q", m["content"])
	}
}

// ── GetConfig handler ─────────────────────────────────────────────────────────

func TestGetConfigHandler_MasksAPIKey(t *testing.T) {
	opts := &mockOptionService{values: map[string]string{
		property.AIProvider.KeyValue: "openai",
		property.AIAPIKey.KeyValue:   "sk-abcdefghijklmnop",
		property.AIModel.KeyValue:    "gpt-4o",
		property.AIBaseURL.KeyValue:  "",
	}}
	h := newHandler(&mockContentService{}, opts)
	ctx := newMockCtx(`{}`)

	resp, err := h.GetConfig(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg, ok := resp.(*aiConfigResponse)
	if !ok {
		t.Fatalf("expected *aiConfigResponse, got %T", resp)
	}
	if cfg.Provider != "openai" {
		t.Errorf("expected provider=openai, got %q", cfg.Provider)
	}
	if cfg.Model != "gpt-4o" {
		t.Errorf("expected model=gpt-4o, got %q", cfg.Model)
	}
	if strings.Contains(cfg.APIKey, "abcdefghijklmno") {
		t.Errorf("API key should be masked, got %q", cfg.APIKey)
	}
	if !strings.Contains(cfg.APIKey, "*") {
		t.Errorf("API key should contain asterisks, got %q", cfg.APIKey)
	}
}

// ── maskAPIKey ────────────────────────────────────────────────────────────────

func TestMaskAPIKey_ShortKey_FullyMasked(t *testing.T) {
	got := maskAPIKey("abc")
	if got != "***" {
		t.Errorf("expected '***', got %q", got)
	}
}

func TestMaskAPIKey_LongKey_PreservesEnds(t *testing.T) {
	got := maskAPIKey("sk-abcdefghijklmnop")
	if !strings.HasPrefix(got, "sk-a") {
		t.Errorf("expected prefix 'sk-a', got %q", got)
	}
	if !strings.HasSuffix(got, "mnop") {
		t.Errorf("expected suffix 'mnop', got %q", got)
	}
	if !strings.Contains(got, "***") {
		t.Errorf("expected middle masked with ***, got %q", got)
	}
}
