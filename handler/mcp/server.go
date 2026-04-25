// Package mcp implements a Model Context Protocol (MCP) server over JSON-RPC 2.0 / HTTP.
// Spec reference: https://spec.modelcontextprotocol.io (protocol version 2024-11-05)
package mcp

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/go-sonic/sonic/cache"
	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/dal/vector"
	"github.com/go-sonic/sonic/handler/web"
	"github.com/go-sonic/sonic/log"
	"github.com/go-sonic/sonic/service"
	ai "github.com/go-sonic/sonic/service/ai"
	"gorm.io/gorm"
)

const mcpProtocolVersion = "2024-11-05"

// Handler is the MCP JSON-RPC 2.0 handler.
type Handler struct {
	postService       service.PostService
	tagService        service.TagService
	embeddingProvider ai.EmbeddingProvider
	vectorStore       ai.VectorStore
	cache             cache.Cache
}

func NewHandler(
	db *gorm.DB,
	postService service.PostService,
	tagService service.TagService,
	embeddingProvider ai.EmbeddingProvider,
	c cache.Cache,
) *Handler {
	return &Handler{
		postService:       postService,
		tagService:        tagService,
		embeddingProvider: embeddingProvider,
		vectorStore:       vector.NewDBStore(db),
		cache:             c,
	}
}

// ── JSON-RPC 2.0 types ──────────────────────────────────────────────────────

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func errResp(id json.RawMessage, code int, msg string) *rpcResponse {
	return &rpcResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: msg}}
}

func okResp(id json.RawMessage, result interface{}) *rpcResponse {
	return &rpcResponse{JSONRPC: "2.0", ID: id, Result: result}
}

// ── Auth ────────────────────────────────────────────────────────────────────

func (h *Handler) authenticate(ctx web.Context) bool {
	token := ctx.Header(consts.AdminTokenHeaderName)
	if token == "" {
		// Also accept Authorization: Bearer <token>
		auth := ctx.Header("Authorization")
		if len(auth) > 7 && auth[:7] == "Bearer " {
			token = auth[7:]
		}
	}
	if token == "" {
		return false
	}
	userID, ok := h.cache.Get(cache.BuildTokenAccessKey(token))
	return ok && userID != nil
}

// ── Main HTTP handler ───────────────────────────────────────────────────────

// Handle processes a single JSON-RPC 2.0 request or a batch.
func (h *Handler) Handle(ctx web.Context) {
	if !h.authenticate(ctx) {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var raw json.RawMessage
	if err := ctx.BindJSON(&raw); err != nil {
		ctx.JSON(http.StatusBadRequest, errResp(nil, -32700, "parse error"))
		return
	}

	reqCtx := ctx.RequestContext()

	// Detect batch vs single
	if len(raw) > 0 && raw[0] == '[' {
		var reqs []rpcRequest
		if err := json.Unmarshal(raw, &reqs); err != nil {
			ctx.JSON(http.StatusOK, errResp(nil, -32700, "parse error"))
			return
		}
		responses := make([]*rpcResponse, 0, len(reqs))
		for _, req := range reqs {
			responses = append(responses, h.dispatch(reqCtx, &req))
		}
		ctx.JSON(http.StatusOK, responses)
		return
	}

	var req rpcRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		ctx.JSON(http.StatusOK, errResp(nil, -32700, "parse error"))
		return
	}
	resp := h.dispatch(reqCtx, &req)
	if resp != nil {
		ctx.JSON(http.StatusOK, resp)
	}
}

func (h *Handler) dispatch(ctx interface{ Value(any) any }, req *rpcRequest) *rpcResponse {
	// Notifications (no id) — process but return nothing
	isNotification := req.ID == nil || string(req.ID) == "null"

	var resp *rpcResponse
	switch req.Method {
	case "initialize":
		resp = h.handleInitialize(req)
	case "initialized":
		return nil // client notification, no response
	case "ping":
		resp = okResp(req.ID, map[string]interface{}{})
	case "tools/list":
		resp = h.handleToolsList(req)
	case "tools/call":
		resp = h.handleToolsCall(req, ctx)
	case "resources/list":
		resp = h.handleResourcesList(req)
	case "resources/read":
		resp = h.handleResourcesRead(req, ctx)
	default:
		if !isNotification {
			resp = errResp(req.ID, -32601, "method not found: "+req.Method)
		}
	}

	if isNotification {
		return nil
	}
	if resp == nil {
		log.Error("mcp: nil response for non-notification", zap.String("method", req.Method))
		resp = errResp(req.ID, -32603, "internal error")
	}
	return resp
}

// ── initialize ──────────────────────────────────────────────────────────────

type initParams struct {
	ProtocolVersion string `json:"protocolVersion"`
	ClientInfo      struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"clientInfo"`
}

func (h *Handler) handleInitialize(req *rpcRequest) *rpcResponse {
	return okResp(req.ID, map[string]interface{}{
		"protocolVersion": mcpProtocolVersion,
		"capabilities": map[string]interface{}{
			"tools":     map[string]interface{}{},
			"resources": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "sonic-evo",
			"version": "1.0.0",
		},
	})
}
