package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	hertzapp "github.com/cloudwego/hertz/pkg/app"

	"github.com/evo-lee/evo-sonic/handler/web/hertzadapter"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

func TestErrorCodeFromStatus(t *testing.T) {
	cases := []struct {
		status int
		code   string
	}{{http.StatusBadRequest, "bad_request"}, {http.StatusUnauthorized, "unauthorized"}, {http.StatusForbidden, "forbidden"}, {http.StatusNotFound, "not_found"}, {http.StatusInternalServerError, "internal_error"}}

	for _, tc := range cases {
		if got := ErrorCodeFromStatus(tc.status); got != tc.code {
			t.Fatalf("status=%d expected code=%s, got=%s", tc.status, tc.code, got)
		}
	}
}

func TestErrorCodeFromError(t *testing.T) {
	cases := []struct {
		err  error
		code string
	}{{xerr.BadParam.New("bad"), "bad_request"}, {xerr.NoRecord.New("nf"), "not_found"}, {xerr.Forbidden.New("forbidden"), "forbidden"}, {xerr.DB.New("db"), "db_error"}, {xerr.Email.New("email"), "email_error"}, {xerr.WithStatus(nil, http.StatusUnauthorized), "unauthorized"}}

	for _, tc := range cases {
		if got := ErrorCodeFromError(tc.err); got != tc.code {
			t.Fatalf("expected code=%s, got=%s", tc.code, got)
		}
	}
}

func TestAbortWithErrorJSONIncludesRequestIDAndCode(t *testing.T) {
	var reqCtx hertzapp.RequestContext
	reqCtx.Request.SetRequestURI("/")
	reqCtx.Request.Header.SetMethod(http.MethodGet)
	reqCtx.Request.Header.Set(RequestIDHeader, "req-test")
	webCtx := hertzadapter.NewContext(context.Background(), &reqCtx)

	NewRequestIDMiddleware().Handler()(webCtx)
	AbortWithErrorJSON(webCtx, http.StatusBadRequest, "bad_request", "bad request")

	if reqCtx.Response.StatusCode() != http.StatusBadRequest {
		t.Fatalf("expected status=%d, got=%d", http.StatusBadRequest, reqCtx.Response.StatusCode())
	}

	var payload map[string]any
	if err := json.Unmarshal(reqCtx.Response.Body(), &payload); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if payload["code"] != "bad_request" {
		t.Fatalf("expected code=bad_request, got=%v", payload["code"])
	}
	if payload["request_id"] != "req-test" {
		t.Fatalf("expected request_id=req-test, got=%v", payload["request_id"])
	}
	if payload["message"] != "bad request" {
		t.Fatalf("expected message=bad request, got=%v", payload["message"])
	}
}
