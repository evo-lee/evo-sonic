package aiimpl

import (
	"context"

	"github.com/evo-lee/evo-sonic/model/property"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/service/ai"
)

// maxTokensCap is the hard upper limit for any single AI request.
// This prevents runaway costs regardless of what callers specify.
const maxTokensCap = 4096

// configurableProvider reads AI settings from the options DB on each call,
// delegating to the appropriate backend provider. It falls back to the YAML
// config (passed at construction) when no DB value is set.
type configurableProvider struct {
	optionService service.OptionService
	yamlAPIKey    string
	yamlModel     string
}

// NewConfigurableProvider is the FX constructor for ai.Provider.
// It replaces the direct Anthropic provider so that runtime config changes
// (via /api/admin/ai/config) take effect without a restart.
func NewConfigurableProvider(optionService service.OptionService) ai.Provider {
	return &configurableProvider{
		optionService: optionService,
	}
}

// resolve builds a concrete provider from the current DB/YAML config.
func (p *configurableProvider) resolve(ctx context.Context) ai.Provider {
	providerName := p.optionService.GetOrByDefault(ctx, property.AIProvider).(string)
	apiKey := p.optionService.GetOrByDefault(ctx, property.AIAPIKey).(string)
	model := p.optionService.GetOrByDefault(ctx, property.AIModel).(string)
	baseURL := p.optionService.GetOrByDefault(ctx, property.AIBaseURL).(string)

	if apiKey == "" {
		return noopProvider{}
	}

	switch providerName {
	case "openai", "ollama":
		return newOpenAIProvider(apiKey, model, baseURL, providerName)
	default:
		return newAnthropicProvider(apiKey, model)
	}
}

func (p *configurableProvider) Complete(ctx context.Context, req ai.CompletionRequest) (ai.CompletionResponse, error) {
	req = capTokens(req)
	return p.resolve(ctx).Complete(ctx, req)
}

func (p *configurableProvider) Stream(ctx context.Context, req ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	req = capTokens(req)
	return p.resolve(ctx).Stream(ctx, req)
}

// capTokens enforces the hard MaxTokens ceiling on any request.
func capTokens(req ai.CompletionRequest) ai.CompletionRequest {
	if req.MaxTokens == 0 || req.MaxTokens > maxTokensCap {
		req.MaxTokens = maxTokensCap
	}
	return req
}
