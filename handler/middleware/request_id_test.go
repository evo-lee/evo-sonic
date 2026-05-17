package middleware

import (
	"context"
	"net/http"
	"testing"

	hertzapp "github.com/cloudwego/hertz/pkg/app"

	"github.com/evo-lee/evo-sonic/handler/web/hertzadapter"
)

func TestRequestIDGenerateWhenHeaderMissing(t *testing.T) {
	var reqCtx hertzapp.RequestContext
	reqCtx.Request.SetRequestURI("/")
	reqCtx.Request.Header.SetMethod(http.MethodGet)
	webCtx := hertzadapter.NewContext(context.Background(), &reqCtx)

	NewRequestIDMiddleware().Handler()(webCtx)

	got := string(reqCtx.Response.Header.Peek(RequestIDHeader))
	if got == "" {
		t.Fatal("expected generated request id in response header")
	}
	if GetRequestID(webCtx) != got {
		t.Fatalf("expected context request id %q, got %q", got, GetRequestID(webCtx))
	}
}

func TestRequestIDUseIncomingHeader(t *testing.T) {
	var reqCtx hertzapp.RequestContext
	reqCtx.Request.SetRequestURI("/")
	reqCtx.Request.Header.SetMethod(http.MethodGet)
	reqCtx.Request.Header.Set(RequestIDHeader, "req-123")
	webCtx := hertzadapter.NewContext(context.Background(), &reqCtx)

	NewRequestIDMiddleware().Handler()(webCtx)

	if got := string(reqCtx.Response.Header.Peek(RequestIDHeader)); got != "req-123" {
		t.Fatalf("expected response request id req-123, got %q", got)
	}
	if got := GetRequestID(webCtx); got != "req-123" {
		t.Fatalf("expected context request id req-123, got %q", got)
	}
}
