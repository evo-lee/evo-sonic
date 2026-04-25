package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-sonic/sonic/consts"
	"github.com/go-sonic/sonic/model/param"
)

// ── Tool schema types ───────────────────────────────────────────────────────

type toolDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema inputSchema `json:"inputSchema"`
}

type inputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

type property struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type toolResult struct {
	Content []contentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func textResult(text string) *toolResult {
	return &toolResult{Content: []contentBlock{{Type: "text", Text: text}}}
}

func errResult(msg string) *toolResult {
	return &toolResult{IsError: true, Content: []contentBlock{{Type: "text", Text: msg}}}
}

func jsonResult(v interface{}) *toolResult {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return errResult("json marshal error: " + err.Error())
	}
	return textResult(string(b))
}

// ── Tool definitions ────────────────────────────────────────────────────────

func toolDefinitions() []toolDef {
	return []toolDef{
		{
			Name:        "list_posts",
			Description: "List blog posts with optional status filter and pagination.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]property{
					"status": {Type: "string", Description: "Filter by status: published, draft, intimate, recycle"},
					"page":   {Type: "integer", Description: "Page number (1-based, default 1)"},
					"size":   {Type: "integer", Description: "Page size (default 10, max 50)"},
				},
			},
		},
		{
			Name:        "get_post",
			Description: "Get full post content including markdown source.",
			InputSchema: inputSchema{
				Type:       "object",
				Properties: map[string]property{"postId": {Type: "integer", Description: "Post ID"}},
				Required:   []string{"postId"},
			},
		},
		{
			Name:        "create_post",
			Description: "Create a new draft post.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]property{
					"title":   {Type: "string", Description: "Post title"},
					"content": {Type: "string", Description: "Post content in Markdown"},
					"summary": {Type: "string", Description: "Short summary (optional)"},
					"tagIds":  {Type: "string", Description: "Comma-separated tag IDs (optional)"},
				},
				Required: []string{"title", "content"},
			},
		},
		{
			Name:        "update_post",
			Description: "Update post content and/or title.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]property{
					"postId":  {Type: "integer", Description: "Post ID"},
					"title":   {Type: "string", Description: "New title (optional)"},
					"content": {Type: "string", Description: "New content (optional)"},
					"summary": {Type: "string", Description: "New summary (optional)"},
				},
				Required: []string{"postId"},
			},
		},
		{
			Name:        "publish_post",
			Description: "Publish a draft post.",
			InputSchema: inputSchema{
				Type:       "object",
				Properties: map[string]property{"postId": {Type: "integer", Description: "Post ID"}},
				Required:   []string{"postId"},
			},
		},
		{
			Name:        "search_posts",
			Description: "Search posts by keyword (full-text title match).",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]property{
					"keyword": {Type: "string", Description: "Search keyword"},
					"size":    {Type: "integer", Description: "Max results (default 10, max 50)"},
				},
				Required: []string{"keyword"},
			},
		},
		{
			Name:        "list_tags",
			Description: "List all tags.",
			InputSchema: inputSchema{Type: "object", Properties: map[string]property{}},
		},
	}
}

// ── handleToolsList ─────────────────────────────────────────────────────────

func (h *Handler) handleToolsList(req *rpcRequest) *rpcResponse {
	return okResp(req.ID, map[string]interface{}{"tools": toolDefinitions()})
}

// ── handleToolsCall ─────────────────────────────────────────────────────────

func (h *Handler) handleToolsCall(req *rpcRequest, ctx interface{ Value(any) any }) *rpcResponse {
	var p toolCallParams
	if err := json.Unmarshal(req.Params, &p); err != nil {
		return errResp(req.ID, -32602, "invalid params")
	}

	goCtx, ok := ctx.(context.Context)
	if !ok {
		goCtx = context.Background()
	}

	var result *toolResult
	switch p.Name {
	case "list_posts":
		result = h.toolListPosts(goCtx, p.Arguments)
	case "get_post":
		result = h.toolGetPost(goCtx, p.Arguments)
	case "create_post":
		result = h.toolCreatePost(goCtx, p.Arguments)
	case "update_post":
		result = h.toolUpdatePost(goCtx, p.Arguments)
	case "publish_post":
		result = h.toolPublishPost(goCtx, p.Arguments)
	case "search_posts":
		result = h.toolSearchPosts(goCtx, p.Arguments)
	case "list_tags":
		result = h.toolListTags(goCtx)
	default:
		return errResp(req.ID, -32602, "unknown tool: "+p.Name)
	}
	return okResp(req.ID, result)
}

// ── Tool implementations ────────────────────────────────────────────────────

func (h *Handler) toolListPosts(ctx context.Context, args json.RawMessage) *toolResult {
	var a struct {
		Status string `json:"status"`
		Page   int    `json:"page"`
		Size   int    `json:"size"`
	}
	_ = json.Unmarshal(args, &a)
	if a.Page < 1 {
		a.Page = 1
	}
	if a.Size < 1 || a.Size > 50 {
		a.Size = 10
	}

	var statuses []*consts.PostStatus
	if a.Status != "" {
		s := parsePostStatus(a.Status)
		statuses = []*consts.PostStatus{&s}
	}

	posts, total, err := h.postService.Page(ctx, param.PostQuery{
		Page:     param.Page{PageNum: a.Page - 1, PageSize: a.Size},
		Statuses: statuses,
	})
	if err != nil {
		return errResult("list_posts error: " + err.Error())
	}

	type postItem struct {
		ID      int32  `json:"id"`
		Title   string `json:"title"`
		Slug    string `json:"slug"`
		Status  string `json:"status"`
		Summary string `json:"summary"`
	}
	items := make([]postItem, 0, len(posts))
	for _, p := range posts {
		items = append(items, postItem{
			ID:      p.ID,
			Title:   p.Title,
			Slug:    p.Slug,
			Status:  postStatusString(p.Status),
			Summary: p.Summary,
		})
	}
	return jsonResult(map[string]interface{}{"total": total, "page": a.Page, "posts": items})
}

func (h *Handler) toolGetPost(ctx context.Context, args json.RawMessage) *toolResult {
	var a struct {
		PostID int32 `json:"postId"`
	}
	if err := json.Unmarshal(args, &a); err != nil || a.PostID == 0 {
		return errResult("postId required")
	}
	post, err := h.postService.GetByPostID(ctx, a.PostID)
	if err != nil {
		return errResult("get_post error: " + err.Error())
	}
	return jsonResult(map[string]interface{}{
		"id":              post.ID,
		"title":           post.Title,
		"slug":            post.Slug,
		"status":          postStatusString(post.Status),
		"summary":         post.Summary,
		"originalContent": post.OriginalContent,
		"createTime":      post.CreateTime,
	})
}

func (h *Handler) toolCreatePost(ctx context.Context, args json.RawMessage) *toolResult {
	var a struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Summary string `json:"summary"`
	}
	if err := json.Unmarshal(args, &a); err != nil || a.Title == "" || a.Content == "" {
		return errResult("title and content required")
	}
	draft := consts.PostStatusDraft
	post, err := h.postService.Create(ctx, &param.Post{
		Title:           a.Title,
		OriginalContent: a.Content,
		Content:         a.Content,
		Summary:         a.Summary,
		Status:          draft,
	})
	if err != nil {
		return errResult("create_post error: " + err.Error())
	}
	return textResult(fmt.Sprintf("Created post id=%d title=%q", post.ID, post.Title))
}

func (h *Handler) toolUpdatePost(ctx context.Context, args json.RawMessage) *toolResult {
	var a struct {
		PostID  int32  `json:"postId"`
		Title   string `json:"title"`
		Content string `json:"content"`
		Summary string `json:"summary"`
	}
	if err := json.Unmarshal(args, &a); err != nil || a.PostID == 0 {
		return errResult("postId required")
	}
	existing, err := h.postService.GetByPostID(ctx, a.PostID)
	if err != nil {
		return errResult("get post error: " + err.Error())
	}
	p := &param.Post{
		Title:           existing.Title,
		OriginalContent: existing.OriginalContent,
		Content:         existing.FormatContent,
		Summary:         existing.Summary,
		Status:          existing.Status,
	}
	if a.Title != "" {
		p.Title = a.Title
	}
	if a.Content != "" {
		p.OriginalContent = a.Content
		p.Content = a.Content
	}
	if a.Summary != "" {
		p.Summary = a.Summary
	}
	_, err = h.postService.Update(ctx, a.PostID, p)
	if err != nil {
		return errResult("update_post error: " + err.Error())
	}
	return textResult(fmt.Sprintf("Updated post id=%d", a.PostID))
}

func (h *Handler) toolPublishPost(ctx context.Context, args json.RawMessage) *toolResult {
	var a struct {
		PostID int32 `json:"postId"`
	}
	if err := json.Unmarshal(args, &a); err != nil || a.PostID == 0 {
		return errResult("postId required")
	}
	_, err := h.postService.UpdateStatus(ctx, a.PostID, consts.PostStatusPublished)
	if err != nil {
		return errResult("publish_post error: " + err.Error())
	}
	return textResult(fmt.Sprintf("Published post id=%d", a.PostID))
}

func (h *Handler) toolSearchPosts(ctx context.Context, args json.RawMessage) *toolResult {
	var a struct {
		Keyword string `json:"keyword"`
		Size    int    `json:"size"`
	}
	if err := json.Unmarshal(args, &a); err != nil || a.Keyword == "" {
		return errResult("keyword required")
	}
	if a.Size < 1 || a.Size > 50 {
		a.Size = 10
	}
	posts, _, err := h.postService.Page(ctx, param.PostQuery{
		Page:    param.Page{PageNum: 0, PageSize: a.Size},
		Keyword: &a.Keyword,
	})
	if err != nil {
		return errResult("search error: " + err.Error())
	}
	type item struct {
		ID    int32  `json:"id"`
		Title string `json:"title"`
		Slug  string `json:"slug"`
	}
	items := make([]item, 0, len(posts))
	for _, p := range posts {
		if strings.Contains(strings.ToLower(p.Title), strings.ToLower(a.Keyword)) ||
			strings.Contains(strings.ToLower(p.OriginalContent), strings.ToLower(a.Keyword)) {
			items = append(items, item{ID: p.ID, Title: p.Title, Slug: p.Slug})
		}
	}
	return jsonResult(map[string]interface{}{"posts": items})
}

func (h *Handler) toolListTags(ctx context.Context) *toolResult {
	tags, err := h.tagService.ListAll(ctx, nil)
	if err != nil {
		return errResult("list_tags error: " + err.Error())
	}
	type tagItem struct {
		ID   int32  `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	items := make([]tagItem, 0, len(tags))
	for _, t := range tags {
		items = append(items, tagItem{ID: t.ID, Name: t.Name, Slug: t.Slug})
	}
	return jsonResult(map[string]interface{}{"tags": items})
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func postStatusString(s consts.PostStatus) string {
	b, _ := s.MarshalJSON()
	if len(b) >= 2 {
		return string(b[1 : len(b)-1]) // strip quotes
	}
	return "UNKNOWN"
}

func parsePostStatus(s string) consts.PostStatus {
	switch strings.ToLower(s) {
	case "published":
		return consts.PostStatusPublished
	case "draft":
		return consts.PostStatusDraft
	case "recycle":
		return consts.PostStatusRecycle
	case "intimate":
		return consts.PostStatusIntimate
	default:
		return consts.PostStatusPublished
	}
}
