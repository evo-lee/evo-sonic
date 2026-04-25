package ai

import "context"

// CompletionRequest is a request to generate text from a prompt.
type CompletionRequest struct {
	System string
	Prompt string
	// MaxTokens defaults to 1024 if zero.
	MaxTokens int
}

// CompletionResponse holds the generated text and token usage.
type CompletionResponse struct {
	Content      string
	InputTokens  int
	OutputTokens int
}

// StreamChunk is a single text fragment from a streaming completion.
type StreamChunk struct {
	Text string
	Err  error
}

// Provider is the abstraction over any LLM backend.
type Provider interface {
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
	// Stream returns a channel of text fragments. The channel is closed when the
	// stream ends or an error occurs. A final chunk with Err != nil signals failure.
	Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)
}
