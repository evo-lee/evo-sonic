package api

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
	"time"

	"github.com/evo-lee/evo-sonic/util/xerr"
)

type semanticSearchMockCtx struct {
	query map[string]string
}

func (c semanticSearchMockCtx) Deadline() (time.Time, bool)                               { return time.Time{}, false }
func (c semanticSearchMockCtx) Done() <-chan struct{}                                     { return nil }
func (c semanticSearchMockCtx) Err() error                                                { return nil }
func (c semanticSearchMockCtx) Value(any) any                                             { return nil }
func (c semanticSearchMockCtx) RequestContext() context.Context                           { return context.Background() }
func (c semanticSearchMockCtx) Request() *http.Request                                    { return nil }
func (c semanticSearchMockCtx) Writer() io.Writer                                         { return io.Discard }
func (c semanticSearchMockCtx) Method() string                                            { return http.MethodGet }
func (c semanticSearchMockCtx) Path() string                                              { return "/api/content/search/semantic" }
func (c semanticSearchMockCtx) RawQuery() string                                          { return "" }
func (c semanticSearchMockCtx) ClientIP() string                                          { return "127.0.0.1" }
func (c semanticSearchMockCtx) Header(string) string                                      { return "" }
func (c semanticSearchMockCtx) ResponseHeader(string) string                              { return "" }
func (c semanticSearchMockCtx) SetHeader(string, string)                                  {}
func (c semanticSearchMockCtx) StatusCode() int                                           { return 0 }
func (c semanticSearchMockCtx) Query(key string) (string, bool)                           { v, ok := c.query[key]; return v, ok }
func (c semanticSearchMockCtx) Param(string) string                                       { return "" }
func (c semanticSearchMockCtx) Cookie(string) (string, error)                             { return "", nil }
func (c semanticSearchMockCtx) SetCookie(string, string, int, string, string, bool, bool) {}
func (c semanticSearchMockCtx) MultipartForm() (*multipart.Form, error)                   { return nil, nil }
func (c semanticSearchMockCtx) Set(string, any)                                           {}
func (c semanticSearchMockCtx) Get(any) (any, bool)                                       { return nil, false }
func (c semanticSearchMockCtx) Bind(any) error                                            { return nil }
func (c semanticSearchMockCtx) BindJSON(any) error                                        { return nil }
func (c semanticSearchMockCtx) BindQuery(any) error                                       { return nil }
func (c semanticSearchMockCtx) BindWith(any, any) error                                   { return nil }
func (c semanticSearchMockCtx) JSON(int, any)                                             {}
func (c semanticSearchMockCtx) AbortWithStatusJSON(int, any)                              {}
func (c semanticSearchMockCtx) String(int, string)                                        {}
func (c semanticSearchMockCtx) Status(int)                                                {}
func (c semanticSearchMockCtx) Redirect(int, string)                                      {}
func (c semanticSearchMockCtx) File(string)                                               {}
func (c semanticSearchMockCtx) FormFile(string) (*multipart.FileHeader, error)            { return nil, nil }
func (c semanticSearchMockCtx) Abort()                                                    {}
func (c semanticSearchMockCtx) Next()                                                     {}
func (c semanticSearchMockCtx) Native() any                                               { return nil }

func TestSemanticSearchMissingQueryReturnsBadRequest(t *testing.T) {
	// Regression: ISSUE-001 - missing semantic search query returned HTTP 500.
	// Found by /qa on 2026-04-25.
	h := &SemanticSearchHandler{}

	_, err := h.Search(semanticSearchMockCtx{query: map[string]string{}})
	if err == nil {
		t.Fatal("expected missing q error")
	}
	if got := xerr.GetHTTPStatus(err); got != xerr.StatusBadRequest {
		t.Fatalf("expected status=%d, got=%d", xerr.StatusBadRequest, got)
	}
	if got := xerr.GetMessage(err); got != "q is required" {
		t.Fatalf("expected message=%q, got=%q", "q is required", got)
	}
}
