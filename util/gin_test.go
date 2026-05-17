package util

import (
	"context"
	"testing"

	hertzapp "github.com/cloudwego/hertz/pkg/app"

	"github.com/evo-lee/evo-sonic/handler/web/hertzadapter"
)

func TestRequestMetadataFromContext(t *testing.T) {
	var reqCtx hertzapp.RequestContext
	reqCtx.Request.Header.Set("User-Agent", "sonic-test-agent")
	reqCtx.Request.Header.Set("X-Forwarded-For", "203.0.113.9")

	webCtx := hertzadapter.NewContext(context.Background(), &reqCtx)
	reqCtxDerived := webCtx.RequestContext()

	if got := GetClientIP(reqCtxDerived); got != "203.0.113.9" {
		t.Fatalf("expected client ip 203.0.113.9, got %q", got)
	}
	if got := GetUserAgent(reqCtxDerived); got != "sonic-test-agent" {
		t.Fatalf("expected user agent sonic-test-agent, got %q", got)
	}
}
