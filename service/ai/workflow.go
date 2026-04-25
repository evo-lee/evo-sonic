package ai

import "context"

// SEOSuggestion is one actionable SEO recommendation for a post.
type SEOSuggestion struct {
	Field      string `json:"field"`
	Severity   string `json:"severity"` // "error" | "warning" | "info"
	Message    string `json:"message"`
}

// SEOReport contains all suggestions for a post.
type SEOReport struct {
	PostID      int32            `json:"post_id"`
	Score       int              `json:"score"` // 0-100
	Suggestions []SEOSuggestion  `json:"suggestions"`
}

// WorkflowService provides AI-driven post quality and SEO analysis.
type WorkflowService interface {
	// CheckSEO validates title, summary, tags, and content and returns actionable suggestions.
	CheckSEO(ctx context.Context, req SEOCheckRequest) (*SEOReport, error)
}

// SEOCheckRequest contains post fields to evaluate.
type SEOCheckRequest struct {
	PostID          int32
	Title           string
	Summary         string
	OriginalContent string
	MetaDescription string
	TagCount        int
}
