package aiimpl

import (
	"context"
	"errors"
	"testing"

	"github.com/go-sonic/sonic/service/ai"
)

// mockProvider is a test double for ai.Provider.
type mockProvider struct {
	completeResp ai.CompletionResponse
	completeErr  error
	streamCh     chan ai.StreamChunk
	streamErr    error
	// capture last request for assertion
	lastReq ai.CompletionRequest
}

func (m *mockProvider) Complete(_ context.Context, req ai.CompletionRequest) (ai.CompletionResponse, error) {
	m.lastReq = req
	return m.completeResp, m.completeErr
}

func (m *mockProvider) Stream(_ context.Context, req ai.CompletionRequest) (<-chan ai.StreamChunk, error) {
	m.lastReq = req
	if m.streamErr != nil {
		return nil, m.streamErr
	}
	if m.streamCh != nil {
		return m.streamCh, nil
	}
	ch := make(chan ai.StreamChunk)
	close(ch)
	return ch, nil
}

func newContentSvc(p ai.Provider) ai.ContentService {
	return NewContentService(p)
}

// ── Summarize ────────────────────────────────────────────────────────────────

func TestSummarize_ReturnsTrimmmedContent(t *testing.T) {
	p := &mockProvider{completeResp: ai.CompletionResponse{Content: "  summary text  "}}
	svc := newContentSvc(p)

	got, err := svc.Summarize(context.Background(), "article body")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "summary text" {
		t.Errorf("expected trimmed content, got %q", got)
	}
}

func TestSummarize_PropagatesProviderError(t *testing.T) {
	p := &mockProvider{completeErr: errors.New("provider down")}
	svc := newContentSvc(p)

	_, err := svc.Summarize(context.Background(), "body")
	if err == nil || err.Error() != "provider down" {
		t.Errorf("expected provider error, got %v", err)
	}
}

func TestSummarize_UsesCorrectSystemPrompt(t *testing.T) {
	p := &mockProvider{completeResp: ai.CompletionResponse{Content: "ok"}}
	svc := newContentSvc(p)

	_, _ = svc.Summarize(context.Background(), "body")
	if p.lastReq.System != summarizeSystemPrompt {
		t.Errorf("expected summarize system prompt, got %q", p.lastReq.System)
	}
}

// ── SuggestTags ──────────────────────────────────────────────────────────────

func TestSuggestTags_ParsesCSVResponse(t *testing.T) {
	p := &mockProvider{completeResp: ai.CompletionResponse{Content: "go,testing,backend"}}
	svc := newContentSvc(p)

	tags, err := svc.SuggestTags(context.Background(), "title", "content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %d: %v", len(tags), tags)
	}
	if tags[0] != "go" || tags[1] != "testing" || tags[2] != "backend" {
		t.Errorf("unexpected tags: %v", tags)
	}
}

func TestSuggestTags_TrimsWhitespace(t *testing.T) {
	p := &mockProvider{completeResp: ai.CompletionResponse{Content: " go , testing , backend "}}
	svc := newContentSvc(p)

	tags, err := svc.SuggestTags(context.Background(), "title", "content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, tag := range tags {
		if tag != "go" && tag != "testing" && tag != "backend" {
			t.Errorf("expected trimmed tag, got %q", tag)
		}
	}
}

func TestSuggestTags_SkipsEmptySegments(t *testing.T) {
	p := &mockProvider{completeResp: ai.CompletionResponse{Content: "go,,backend,"}}
	svc := newContentSvc(p)

	tags, err := svc.SuggestTags(context.Background(), "title", "content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("expected 2 non-empty tags, got %d: %v", len(tags), tags)
	}
}

func TestSuggestTags_PropagatesProviderError(t *testing.T) {
	p := &mockProvider{completeErr: errors.New("timeout")}
	svc := newContentSvc(p)

	_, err := svc.SuggestTags(context.Background(), "title", "content")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// ── Polish ───────────────────────────────────────────────────────────────────

func TestPolish_ReturnsTrimmmedContent(t *testing.T) {
	p := &mockProvider{completeResp: ai.CompletionResponse{Content: "\npolished\n"}}
	svc := newContentSvc(p)

	got, err := svc.Polish(context.Background(), "draft")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "polished" {
		t.Errorf("expected trimmed content, got %q", got)
	}
}

func TestPolish_SetsMaxTokens(t *testing.T) {
	p := &mockProvider{completeResp: ai.CompletionResponse{Content: "ok"}}
	svc := newContentSvc(p)

	_, _ = svc.Polish(context.Background(), "draft")
	if p.lastReq.MaxTokens != 4096 {
		t.Errorf("expected MaxTokens=4096, got %d", p.lastReq.MaxTokens)
	}
}

// ── Streaming ────────────────────────────────────────────────────────────────

func TestSummarizeStream_ReturnsDelegatedChannel(t *testing.T) {
	ch := make(chan ai.StreamChunk, 1)
	ch <- ai.StreamChunk{Text: "chunk"}
	close(ch)

	p := &mockProvider{streamCh: ch}
	svc := newContentSvc(p)

	got, err := svc.SummarizeStream(context.Background(), "body")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	chunk := <-got
	if chunk.Text != "chunk" {
		t.Errorf("expected 'chunk', got %q", chunk.Text)
	}
}

func TestPolishStream_ReturnsDelegatedChannel(t *testing.T) {
	ch := make(chan ai.StreamChunk, 1)
	ch <- ai.StreamChunk{Text: "polish"}
	close(ch)

	p := &mockProvider{streamCh: ch}
	svc := newContentSvc(p)

	got, err := svc.PolishStream(context.Background(), "draft")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	chunk := <-got
	if chunk.Text != "polish" {
		t.Errorf("expected 'polish', got %q", chunk.Text)
	}
}

func TestSuggestTagsStream_ReturnsDelegatedChannel(t *testing.T) {
	ch := make(chan ai.StreamChunk, 1)
	ch <- ai.StreamChunk{Text: "go,backend"}
	close(ch)

	p := &mockProvider{streamCh: ch}
	svc := newContentSvc(p)

	got, err := svc.SuggestTagsStream(context.Background(), "title", "body")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	chunk := <-got
	if chunk.Text != "go,backend" {
		t.Errorf("expected 'go,backend', got %q", chunk.Text)
	}
}

func TestSuggestTagsStream_PropagatesStreamError(t *testing.T) {
	p := &mockProvider{streamErr: errors.New("stream init failed")}
	svc := newContentSvc(p)

	_, err := svc.SuggestTagsStream(context.Background(), "title", "body")
	if err == nil || err.Error() != "stream init failed" {
		t.Errorf("expected stream error, got %v", err)
	}
}
