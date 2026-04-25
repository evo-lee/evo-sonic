package aiimpl

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/go-sonic/sonic/service/ai"
)

type anthropicProvider struct {
	client anthropic.Client
	model  string
}

func newAnthropicProvider(apiKey, model string) ai.Provider {
	if model == "" {
		model = anthropic.ModelClaudeHaiku4_5_20251001
	}
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &anthropicProvider{client: client, model: model}
}

func (p *anthropicProvider) Complete(ctx context.Context, req ai.CompletionRequest) (ai.CompletionResponse, error) {
	maxTokens := int64(req.MaxTokens)
	if maxTokens == 0 {
		maxTokens = 1024
	}

	params := anthropic.MessageNewParams{
		Model:     p.model,
		MaxTokens: maxTokens,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(req.Prompt)),
		},
	}
	if req.System != "" {
		params.System = []anthropic.TextBlockParam{{Text: req.System}}
	}

	msg, err := p.client.Messages.New(ctx, params)
	if err != nil {
		return ai.CompletionResponse{}, fmt.Errorf("anthropic complete: %w", err)
	}

	var text string
	for _, block := range msg.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}
	return ai.CompletionResponse{
		Content:      text,
		InputTokens:  int(msg.Usage.InputTokens),
		OutputTokens: int(msg.Usage.OutputTokens),
	}, nil
}

func (p *anthropicProvider) Stream(ctx context.Context, req ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	maxTokens := int64(req.MaxTokens)
	if maxTokens == 0 {
		maxTokens = 4096
	}

	params := anthropic.MessageNewParams{
		Model:     p.model,
		MaxTokens: maxTokens,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(req.Prompt)),
		},
	}
	if req.System != "" {
		params.System = []anthropic.TextBlockParam{{Text: req.System}}
	}

	stream := p.client.Messages.NewStreaming(ctx, params)
	ch := make(chan ai.StreamChunk, 16)

	go func() {
		defer close(ch)
		defer stream.Close()
		for stream.Next() {
			event := stream.Current()
			if event.Delta.Type == "text_delta" && event.Delta.Text != "" {
				ch <- ai.StreamChunk{Text: event.Delta.Text}
			}
		}
		if err := stream.Err(); err != nil {
			ch <- ai.StreamChunk{Err: fmt.Errorf("anthropic stream: %w", err)}
		}
	}()

	return ch, nil
}

// noopProvider is returned when no API key is configured.
type noopProvider struct{}

func (noopProvider) Complete(_ context.Context, _ ai.CompletionRequest) (ai.CompletionResponse, error) {
	return ai.CompletionResponse{}, fmt.Errorf("AI not configured: set ai_api_key via /api/admin/ai/config")
}

func (noopProvider) Stream(_ context.Context, _ ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	ch := make(chan ai.StreamChunk, 1)
	ch <- ai.StreamChunk{Err: fmt.Errorf("AI not configured: set ai_api_key via /api/admin/ai/config")}
	close(ch)
	return ch, nil
}
