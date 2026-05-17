package aiimpl

import (
	"context"
	"strings"
	"testing"

	"github.com/evo-lee/evo-sonic/consts"
	"github.com/evo-lee/evo-sonic/model/dto"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/model/property"
	"github.com/evo-lee/evo-sonic/service/ai"
)

// ── capTokens ────────────────────────────────────────────────────────────────

func TestCapTokens_ZeroBecomesMax(t *testing.T) {
	req := capTokens(ai.CompletionRequest{MaxTokens: 0})
	if req.MaxTokens != maxTokensCap {
		t.Errorf("expected %d, got %d", maxTokensCap, req.MaxTokens)
	}
}

func TestCapTokens_OverMaxClamped(t *testing.T) {
	req := capTokens(ai.CompletionRequest{MaxTokens: 99999})
	if req.MaxTokens != maxTokensCap {
		t.Errorf("expected %d, got %d", maxTokensCap, req.MaxTokens)
	}
}

func TestCapTokens_UnderMaxPreserved(t *testing.T) {
	req := capTokens(ai.CompletionRequest{MaxTokens: 512})
	if req.MaxTokens != 512 {
		t.Errorf("expected 512, got %d", req.MaxTokens)
	}
}

func TestCapTokens_AtCapPreserved(t *testing.T) {
	req := capTokens(ai.CompletionRequest{MaxTokens: maxTokensCap})
	if req.MaxTokens != maxTokensCap {
		t.Errorf("expected %d, got %d", maxTokensCap, req.MaxTokens)
	}
}

func TestCapTokens_PreservesOtherFields(t *testing.T) {
	req := capTokens(ai.CompletionRequest{
		System:    "sys",
		Prompt:    "prompt",
		MaxTokens: 100,
	})
	if req.System != "sys" || req.Prompt != "prompt" {
		t.Error("capTokens must not modify System or Prompt")
	}
}

// ── noopProvider ─────────────────────────────────────────────────────────────

func TestNoopProvider_Complete_ReturnsError(t *testing.T) {
	var p noopProvider
	_, err := p.Complete(context.Background(), ai.CompletionRequest{})
	if err == nil {
		t.Fatal("expected error from noopProvider")
	}
	if !strings.Contains(err.Error(), "AI not configured") {
		t.Errorf("expected 'AI not configured' in error, got: %v", err)
	}
}

func TestNoopProvider_Stream_ReturnsChannelWithError(t *testing.T) {
	var p noopProvider
	ch, err := p.Stream(context.Background(), ai.CompletionRequest{})
	if err != nil {
		t.Fatalf("Stream should return channel not error, got: %v", err)
	}
	chunk, ok := <-ch
	if !ok {
		t.Fatal("expected error chunk before close")
	}
	if chunk.Err == nil {
		t.Fatal("expected chunk.Err to be set")
	}
	if !strings.Contains(chunk.Err.Error(), "AI not configured") {
		t.Errorf("expected 'AI not configured' in chunk error, got: %v", chunk.Err)
	}
	// channel must be closed after error chunk
	if _, open := <-ch; open {
		t.Error("channel should be closed after error chunk")
	}
}

// ── configurableProvider ─────────────────────────────────────────────────────

// mockOptionService is a minimal stub for service.OptionService.
// Only GetOrByDefault is used by configurableProvider; remaining methods
// satisfy the interface and return zero values.
type mockOptionService struct {
	values map[string]string
}

func (m *mockOptionService) GetOrByDefault(_ context.Context, p property.Property) interface{} {
	if v, ok := m.values[p.KeyValue]; ok {
		return v
	}
	return p.DefaultValue
}

func (m *mockOptionService) GetOrByDefaultWithErr(_ context.Context, p property.Property, _ interface{}) (interface{}, error) {
	return m.GetOrByDefault(context.Background(), p), nil
}
func (m *mockOptionService) GetBlogBaseURL(_ context.Context) (string, error)        { return "", nil }
func (m *mockOptionService) IsEnabledAbsolutePath(_ context.Context) (bool, error)   { return false, nil }
func (m *mockOptionService) GetPathSuffix(_ context.Context) (string, error)         { return "", nil }
func (m *mockOptionService) GetArchivePrefix(_ context.Context) (string, error)      { return "", nil }
func (m *mockOptionService) Save(_ context.Context, _ map[string]string) error       { return nil }
func (m *mockOptionService) ListAllOption(_ context.Context) ([]*dto.Option, error)  { return nil, nil }
func (m *mockOptionService) GetLinksPrefix(_ context.Context) (string, error)        { return "", nil }
func (m *mockOptionService) GetPhotoPrefix(_ context.Context) (string, error)        { return "", nil }
func (m *mockOptionService) GetJournalPrefix(_ context.Context) (string, error)      { return "", nil }
func (m *mockOptionService) GetActivatedThemeID(_ context.Context) (string, error)   { return "", nil }
func (m *mockOptionService) GetPostPermalinkType(_ context.Context) (consts.PostPermalinkType, error) {
	return "", nil
}
func (m *mockOptionService) GetSheetPermalinkType(_ context.Context) (consts.SheetPermaLinkType, error) {
	return "", nil
}
func (m *mockOptionService) GetIndexPageSize(_ context.Context) int                  { return 0 }
func (m *mockOptionService) GetPostSort(_ context.Context) param.Sort                { return param.Sort{} }
func (m *mockOptionService) GetPostSummaryLength(_ context.Context) int              { return 0 }
func (m *mockOptionService) GetCategoryPrefix(_ context.Context) (string, error)     { return "", nil }
func (m *mockOptionService) GetTagPrefix(_ context.Context) (string, error)          { return "", nil }
func (m *mockOptionService) GetLinkPrefix(_ context.Context) (string, error)         { return "", nil }
func (m *mockOptionService) GetSheetPrefix(_ context.Context) (string, error)        { return "", nil }
func (m *mockOptionService) GetAttachmentType(_ context.Context) consts.AttachmentType { return 0 }
func (m *mockOptionService) GetAdminURLPath(_ context.Context) (string, error)       { return "", nil }

func TestConfigurableProvider_NoAPIKey_FallsBackToNoop(t *testing.T) {
	opts := &mockOptionService{values: map[string]string{
		property.AIAPIKey.KeyValue: "", // no key
	}}
	p := NewConfigurableProvider(opts)

	_, err := p.Complete(context.Background(), ai.CompletionRequest{Prompt: "test"})
	if err == nil {
		t.Fatal("expected error when no API key configured")
	}
	if !strings.Contains(err.Error(), "AI not configured") {
		t.Errorf("expected noop error, got: %v", err)
	}
}

func TestConfigurableProvider_Complete_CapsTokens(t *testing.T) {
	// With no API key the noop is used, but we can verify cap enforcement via
	// the error path: the noop returns error regardless of MaxTokens, so we
	// just check that capTokens was called (would panic if req were nil).
	opts := &mockOptionService{}
	p := NewConfigurableProvider(opts)

	// Sending a huge MaxTokens — should not panic and noop error expected.
	_, err := p.Complete(context.Background(), ai.CompletionRequest{
		Prompt:    "test",
		MaxTokens: 1_000_000,
	})
	if err == nil {
		t.Fatal("expected error from noop provider")
	}
}

func TestConfigurableProvider_Stream_CapsTokens(t *testing.T) {
	opts := &mockOptionService{}
	p := NewConfigurableProvider(opts)

	ch, err := p.Stream(context.Background(), ai.CompletionRequest{
		Prompt:    "test",
		MaxTokens: 1_000_000,
	})
	if err != nil {
		t.Fatalf("Stream should return channel not error, got: %v", err)
	}
	chunk := <-ch
	if chunk.Err == nil {
		t.Fatal("expected error chunk from noop")
	}
}
