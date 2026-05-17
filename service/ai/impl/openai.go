package aiimpl

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	aiservice "github.com/evo-lee/evo-sonic/service/ai"
)

const (
	defaultOllamaBaseURL = "http://localhost:11434/v1"
	defaultOllamaModel   = "llama3"
	defaultOpenAIModel   = "gpt-4o-mini"
)

type openAIProvider struct {
	client       openai.Client
	model        string
	providerName string // "openai" or "ollama"
}

func newOpenAIProvider(apiKey, model, baseURL, providerName string) aiservice.Provider {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	} else if providerName == "ollama" {
		opts = append(opts, option.WithBaseURL(defaultOllamaBaseURL))
	}
	if model == "" {
		if providerName == "ollama" {
			model = defaultOllamaModel
		} else {
			model = defaultOpenAIModel
		}
	}
	client := openai.NewClient(opts...)
	return &openAIProvider{client: client, model: model, providerName: providerName}
}

func (p *openAIProvider) Complete(ctx context.Context, req aiservice.CompletionRequest) (aiservice.CompletionResponse, error) {
	messages := []openai.ChatCompletionMessageParamUnion{}
	if req.System != "" {
		messages = append(messages, openai.SystemMessage(req.System))
	}
	messages = append(messages, openai.UserMessage(req.Prompt))

	maxTokens := int64(req.MaxTokens)
	if maxTokens == 0 {
		maxTokens = 1024
	}

	params := openai.ChatCompletionNewParams{
		Model:     p.model,
		Messages:  messages,
		MaxTokens: openai.Int(maxTokens),
	}

	resp, err := p.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return aiservice.CompletionResponse{}, fmt.Errorf("%s complete: %w", p.providerName, err)
	}

	var text string
	if len(resp.Choices) > 0 {
		text = resp.Choices[0].Message.Content
	}
	return aiservice.CompletionResponse{
		Content:      text,
		InputTokens:  int(resp.Usage.PromptTokens),
		OutputTokens: int(resp.Usage.CompletionTokens),
	}, nil
}

func (p *openAIProvider) Stream(ctx context.Context, req aiservice.CompletionRequest) (<-chan aiservice.StreamChunk, error) {
	messages := []openai.ChatCompletionMessageParamUnion{}
	if req.System != "" {
		messages = append(messages, openai.SystemMessage(req.System))
	}
	messages = append(messages, openai.UserMessage(req.Prompt))

	maxTokens := int64(req.MaxTokens)
	if maxTokens == 0 {
		maxTokens = 4096
	}

	params := openai.ChatCompletionNewParams{
		Model:     p.model,
		Messages:  messages,
		MaxTokens: openai.Int(maxTokens),
	}

	stream := p.client.Chat.Completions.NewStreaming(ctx, params)
	ch := make(chan aiservice.StreamChunk, 16)

	go func() {
		defer close(ch)
		defer stream.Close()
		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) > 0 {
				text := chunk.Choices[0].Delta.Content
				if text != "" {
					ch <- aiservice.StreamChunk{Text: text}
				}
			}
		}
		if err := stream.Err(); err != nil {
			ch <- aiservice.StreamChunk{Err: fmt.Errorf("%s stream: %w", p.providerName, err)}
		}
	}()

	return ch, nil
}
