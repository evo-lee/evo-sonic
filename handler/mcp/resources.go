package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/evo-lee/evo-sonic/consts"
	"github.com/evo-lee/evo-sonic/model/param"
)

// ── Resource schema types ────────────────────────────────────────────────────

type resourceDef struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MIMEType    string `json:"mimeType"`
}

type readResourceParams struct {
	URI string `json:"uri"`
}

// ── handleResourcesList ──────────────────────────────────────────────────────

func (h *Handler) handleResourcesList(req *rpcRequest) *rpcResponse {
	return okResp(req.ID, map[string]interface{}{
		"resources": []resourceDef{
			{URI: "posts://published", Name: "Published Posts", Description: "All published blog posts", MIMEType: "application/json"},
			{URI: "posts://drafts", Name: "Draft Posts", Description: "All draft posts", MIMEType: "application/json"},
			{URI: "tags://all", Name: "Tags", Description: "All tags", MIMEType: "application/json"},
		},
	})
}

// ── handleResourcesRead ──────────────────────────────────────────────────────

type resourceContent struct {
	URI      string `json:"uri"`
	MIMEType string `json:"mimeType"`
	Text     string `json:"text"`
}

func (h *Handler) handleResourcesRead(req *rpcRequest, ctx interface{ Value(any) any }) *rpcResponse {
	var p readResourceParams
	if err := json.Unmarshal(req.Params, &p); err != nil {
		return errResp(req.ID, -32602, "invalid params")
	}

	goCtx, ok := ctx.(context.Context)
	if !ok {
		goCtx = context.Background()
	}

	var content *resourceContent
	var readErr error

	switch p.URI {
	case "posts://published":
		content, readErr = h.readPosts(goCtx, consts.PostStatusPublished, p.URI)
	case "posts://drafts":
		content, readErr = h.readPosts(goCtx, consts.PostStatusDraft, p.URI)
	case "tags://all":
		content, readErr = h.readTags(goCtx, p.URI)
	default:
		return errResp(req.ID, -32602, fmt.Sprintf("unknown resource URI: %s", p.URI))
	}

	if readErr != nil {
		return errResp(req.ID, -32603, readErr.Error())
	}
	return okResp(req.ID, map[string]interface{}{
		"contents": []resourceContent{*content},
	})
}

func (h *Handler) readPosts(ctx context.Context, status consts.PostStatus, uri string) (*resourceContent, error) {
	posts, _, err := h.postService.Page(ctx, param.PostQuery{
		Page:     param.Page{PageNum: 0, PageSize: 100},
		Statuses: []*consts.PostStatus{&status},
	})
	if err != nil {
		return nil, err
	}
	type item struct {
		ID      int32  `json:"id"`
		Title   string `json:"title"`
		Slug    string `json:"slug"`
		Summary string `json:"summary"`
	}
	items := make([]item, 0, len(posts))
	for _, p := range posts {
		items = append(items, item{ID: p.ID, Title: p.Title, Slug: p.Slug, Summary: p.Summary})
	}
	b, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return nil, err
	}
	return &resourceContent{URI: uri, MIMEType: "application/json", Text: string(b)}, nil
}

func (h *Handler) readTags(ctx context.Context, uri string) (*resourceContent, error) {
	tags, err := h.tagService.ListAll(ctx, nil)
	if err != nil {
		return nil, err
	}
	type item struct {
		ID   int32  `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	items := make([]item, 0, len(tags))
	for _, t := range tags {
		items = append(items, item{ID: t.ID, Name: t.Name, Slug: t.Slug})
	}
	b, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return nil, err
	}
	return &resourceContent{URI: uri, MIMEType: "application/json", Text: string(b)}, nil
}
