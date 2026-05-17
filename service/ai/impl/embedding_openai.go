package aiimpl

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	ai "github.com/evo-lee/evo-sonic/service/ai"
)

const defaultEmbeddingModel = "text-embedding-3-small"

type openAIEmbeddingProvider struct {
	client openai.Client
	model  string
}

func newOpenAIEmbeddingProvider(apiKey, model, baseURL string) ai.EmbeddingProvider {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	if model == "" {
		model = defaultEmbeddingModel
	}
	return &openAIEmbeddingProvider{
		client: openai.NewClient(opts...),
		model:  model,
	}
}

func (p *openAIEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	resp, err := p.client.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{OfArrayOfStrings: []string{text}},
		Model: openai.EmbeddingModel(p.model),
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("embedding: empty response from model %s", p.model)
	}
	raw := resp.Data[0].Embedding
	vec := make([]float32, len(raw))
	for i, v := range raw {
		vec[i] = float32(v)
	}
	return vec, nil
}

// noopEmbeddingProvider is returned when embedding is not configured.
type noopEmbeddingProvider struct{}

func (noopEmbeddingProvider) Embed(_ context.Context, _ string) ([]float32, error) {
	return nil, fmt.Errorf("embedding not configured: set ai_provider to openai/ollama with a valid api_key")
}
