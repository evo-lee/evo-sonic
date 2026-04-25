package ai

import "context"

// ContentService provides AI-powered content enhancement.
type ContentService interface {
	// Summarize generates a short summary for the given markdown content.
	Summarize(ctx context.Context, content string) (string, error)

	// SuggestTags returns a list of tag names suitable for the content.
	SuggestTags(ctx context.Context, title, content string) ([]string, error)

	// Polish rewrites the content for clarity and readability.
	Polish(ctx context.Context, content string) (string, error)

	// PolishStream streams the polished content chunk by chunk.
	PolishStream(ctx context.Context, content string) (<-chan StreamChunk, error)

	// SummarizeStream streams the generated summary chunk by chunk.
	SummarizeStream(ctx context.Context, content string) (<-chan StreamChunk, error)

	// SuggestTagsStream streams tag suggestions chunk by chunk.
	SuggestTagsStream(ctx context.Context, title, content string) (<-chan StreamChunk, error)
}
