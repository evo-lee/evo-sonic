package ai

import "context"

// EmbeddingProvider converts text into a dense float vector.
type EmbeddingProvider interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// SearchResult is one item returned from a vector similarity search.
type SearchResult struct {
	PostID int32
	Score  float64
}

// VectorStore persists and searches post embedding vectors.
type VectorStore interface {
	Upsert(ctx context.Context, postID int32, model string, vec []float32) error
	Delete(ctx context.Context, postID int32) error
	// Search returns the topK most similar posts to the query vector, ordered by descending score.
	Search(ctx context.Context, vec []float32, topK int) ([]SearchResult, error)
}
