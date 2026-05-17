package aiimpl

import (
	"context"

	"github.com/evo-lee/evo-sonic/model/property"
	"github.com/evo-lee/evo-sonic/service"
	ai "github.com/evo-lee/evo-sonic/service/ai"
)

// configurableEmbeddingProvider resolves an EmbeddingProvider from DB config on each call,
// matching the pattern used by configurableProvider for LLM text generation.
type configurableEmbeddingProvider struct {
	optionService service.OptionService
}

func NewConfigurableEmbeddingProvider(optionService service.OptionService) ai.EmbeddingProvider {
	return &configurableEmbeddingProvider{optionService: optionService}
}

func (p *configurableEmbeddingProvider) resolve(ctx context.Context) ai.EmbeddingProvider {
	providerName := p.optionService.GetOrByDefault(ctx, property.AIProvider).(string)
	apiKey := p.optionService.GetOrByDefault(ctx, property.AIAPIKey).(string)
	model := p.optionService.GetOrByDefault(ctx, property.AIEmbeddingModel).(string)
	baseURL := p.optionService.GetOrByDefault(ctx, property.AIBaseURL).(string)

	if apiKey == "" {
		return noopEmbeddingProvider{}
	}
	// Anthropic has no embedding API — require openai/ollama or a compatible base_url.
	if providerName == "anthropic" && baseURL == "" {
		return noopEmbeddingProvider{}
	}
	return newOpenAIEmbeddingProvider(apiKey, model, baseURL)
}

func (p *configurableEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	return p.resolve(ctx).Embed(ctx, text)
}
