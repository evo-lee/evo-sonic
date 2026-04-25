package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
	"time"
)

// ── minimal web.Context mock ──────────────────────────────────────────────────

type mockCtx struct {
	body    []byte
	jsonOut interface{}
	jsonCode int
	headers  map[string]string
}

func newCtx(body interface{}) *mockCtx {
	b, _ := json.Marshal(body)
	return &mockCtx{body: b, headers: make(map[string]string)}
}

func (c *mockCtx) BindJSON(v any) error {
	return json.NewDecoder(bytes.NewReader(c.body)).Decode(v)
}
func (c *mockCtx) JSON(code int, v any)                    { c.jsonCode = code; c.jsonOut = v }
func (c *mockCtx) AbortWithStatusJSON(code int, v any)     { c.jsonCode = code; c.jsonOut = v }
func (c *mockCtx) Header(k string) string                   { return c.headers[k] }
func (c *mockCtx) SetHeader(k, v string)                    { c.headers[k] = v }
func (c *mockCtx) RequestContext() context.Context          { return context.Background() }
func (c *mockCtx) Writer() io.Writer                        { return io.Discard }
func (c *mockCtx) Request() *http.Request                   { return nil }
func (c *mockCtx) Method() string                           { return "POST" }
func (c *mockCtx) Path() string                             { return "/mcp/v1" }
func (c *mockCtx) RawQuery() string                         { return "" }
func (c *mockCtx) ClientIP() string                         { return "127.0.0.1" }
func (c *mockCtx) ResponseHeader(_ string) string           { return "" }
func (c *mockCtx) StatusCode() int                          { return c.jsonCode }
func (c *mockCtx) Query(_ string) (string, bool)            { return "", false }
func (c *mockCtx) Param(_ string) string                    { return "" }
func (c *mockCtx) Cookie(_ string) (string, error)          { return "", nil }
func (c *mockCtx) SetCookie(_, _ string, _ int, _, _ string, _, _ bool) {}
func (c *mockCtx) String(_ int, _ string)                   {}
func (c *mockCtx) Status(_ int)                             {}
func (c *mockCtx) Redirect(_ int, _ string)                 {}
func (c *mockCtx) File(_ string)                            {}
func (c *mockCtx) FormFile(_ string) (*multipart.FileHeader, error) { return nil, nil }
func (c *mockCtx) Abort()                                   {}
func (c *mockCtx) Next()                                    {}
func (c *mockCtx) Native() any                              { return nil }
func (c *mockCtx) BindQuery(_ any) error                    { return nil }
func (c *mockCtx) BindWith(_ any, _ any) error              { return nil }
func (c *mockCtx) Bind(_ any) error                         { return nil }
func (c *mockCtx) Set(_ string, _ any)                      {}
func (c *mockCtx) Get(_ any) (any, bool)                    { return nil, false }
func (c *mockCtx) Deadline() (time.Time, bool)              { return time.Time{}, false }
func (c *mockCtx) Done() <-chan struct{}                     { return nil }
func (c *mockCtx) Err() error                               { return nil }
func (c *mockCtx) Value(_ any) any                          { return nil }
func (c *mockCtx) MultipartForm() (*multipart.Form, error)  { return nil, nil }

// ── mock cache ────────────────────────────────────────────────────────────────

type mockCache struct{ authed bool }

func (m *mockCache) Get(key string) (interface{}, bool) {
	if m.authed {
		return int32(1), true
	}
	return nil, false
}
func (m *mockCache) SetDefault(key string, val interface{})          {}
func (m *mockCache) Set(key string, val interface{}, ttl time.Duration) {}
func (m *mockCache) Delete(key string)                               {}
func (m *mockCache) BatchDelete(keys []string)                       {}
func (m *mockCache) BuildTokenAccessKey(token string) string         { return "token:" + token }

// Build token key the same way cache.BuildTokenAccessKey does
// (handler uses that package function — we just need cache.Get to work)

// ── helpers ───────────────────────────────────────────────────────────────────

func newHandler(authed bool) *Handler {
	return &Handler{cache: &mockCache{authed: authed}}
}

func dispatch(h *Handler, req interface{}) *rpcResponse {
	raw, _ := json.Marshal(req)
	var r rpcRequest
	_ = json.Unmarshal(raw, &r)
	return h.dispatch(context.Background(), &r)
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestDispatch_Initialize_ReturnsProtocolVersion(t *testing.T) {
	h := newHandler(true)
	resp := dispatch(h, map[string]interface{}{
		"jsonrpc": "2.0", "id": 1, "method": "initialize",
	})
	if resp == nil || resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp)
	}
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("result not a map")
	}
	if result["protocolVersion"] != mcpProtocolVersion {
		t.Errorf("want %s got %v", mcpProtocolVersion, result["protocolVersion"])
	}
}

func TestDispatch_Ping_ReturnsEmpty(t *testing.T) {
	h := newHandler(true)
	resp := dispatch(h, map[string]interface{}{"jsonrpc": "2.0", "id": 2, "method": "ping"})
	if resp == nil || resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp)
	}
}

func TestDispatch_UnknownMethod_MethodNotFound(t *testing.T) {
	h := newHandler(true)
	resp := dispatch(h, map[string]interface{}{"jsonrpc": "2.0", "id": 3, "method": "unknown/method"})
	if resp == nil || resp.Error == nil {
		t.Fatal("expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("want -32601 got %d", resp.Error.Code)
	}
}

func TestDispatch_Notification_NoResponse(t *testing.T) {
	h := newHandler(true)
	// No id = notification
	resp := dispatch(h, map[string]interface{}{"jsonrpc": "2.0", "method": "initialized"})
	if resp != nil {
		t.Error("notifications must return nil")
	}
}

func TestDispatch_ToolsList_ReturnsTools(t *testing.T) {
	h := newHandler(true)
	resp := dispatch(h, map[string]interface{}{"jsonrpc": "2.0", "id": 4, "method": "tools/list"})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	tools, ok := result["tools"].([]toolDef)
	if !ok {
		t.Fatalf("tools not []toolDef, got %T", result["tools"])
	}
	if len(tools) == 0 {
		t.Error("expected at least one tool")
	}
}

func TestDispatch_ResourcesList_ReturnsResources(t *testing.T) {
	h := newHandler(true)
	resp := dispatch(h, map[string]interface{}{"jsonrpc": "2.0", "id": 5, "method": "resources/list"})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %v", resp.Error)
	}
	result := resp.Result.(map[string]interface{})
	resources, ok := result["resources"].([]resourceDef)
	if !ok {
		t.Fatalf("resources not []resourceDef, got %T", result["resources"])
	}
	if len(resources) == 0 {
		t.Error("expected at least one resource")
	}
}

func TestDispatch_ToolsCall_UnknownTool_Error(t *testing.T) {
	h := newHandler(true)
	resp := dispatch(h, map[string]interface{}{
		"jsonrpc": "2.0", "id": 6, "method": "tools/call",
		"params": map[string]interface{}{
			"name": "does_not_exist", "arguments": map[string]interface{}{},
		},
	})
	if resp.Error == nil {
		t.Fatal("expected error for unknown tool")
	}
	if resp.Error.Code != -32602 {
		t.Errorf("want -32602 got %d", resp.Error.Code)
	}
}

func TestAuthenticate_MissingToken_ReturnsFalse(t *testing.T) {
	h := newHandler(false)
	ctx := newCtx(map[string]string{})
	if h.authenticate(ctx) {
		t.Error("expected false with no token")
	}
}

func TestAuthenticate_BearerToken_Accepted(t *testing.T) {
	h := newHandler(true)
	ctx := newCtx(map[string]string{})
	ctx.headers["Authorization"] = "Bearer valid-token"
	if !h.authenticate(ctx) {
		t.Error("expected true with valid Bearer token")
	}
}
