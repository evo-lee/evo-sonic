package vector

import (
	"math"
	"testing"
)

// ── encodeVector / decodeVector ───────────────────────────────────────────────

func TestEncodeDecodeVector_RoundTrip(t *testing.T) {
	in := []float32{0.1, -0.5, 1.0, 0.0, 3.14}
	b := encodeVector(in)
	out := decodeVector(b)
	if len(out) != len(in) {
		t.Fatalf("length mismatch: want %d got %d", len(in), len(out))
	}
	for i := range in {
		if math.Abs(float64(out[i]-in[i])) > 1e-6 {
			t.Errorf("[%d] want %v got %v", i, in[i], out[i])
		}
	}
}

func TestDecodeVector_EmptyBytes(t *testing.T) {
	if got := decodeVector([]byte{}); len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestDecodeVector_UnalignedBytes(t *testing.T) {
	if got := decodeVector([]byte{1, 2, 3}); got != nil {
		t.Errorf("expected nil for unaligned bytes, got %v", got)
	}
}

func TestEncodeVector_Empty(t *testing.T) {
	b := encodeVector(nil)
	if len(b) != 0 {
		t.Errorf("expected empty, got len=%d", len(b))
	}
}

// ── cosineSimilarity ──────────────────────────────────────────────────────────

func TestCosineSimilarity_IdenticalVectors(t *testing.T) {
	v := []float32{1, 2, 3}
	got := cosineSimilarity(v, v)
	if math.Abs(got-1.0) > 1e-6 {
		t.Errorf("identical vectors: want 1.0 got %v", got)
	}
}

func TestCosineSimilarity_OrthogonalVectors(t *testing.T) {
	a := []float32{1, 0}
	b := []float32{0, 1}
	got := cosineSimilarity(a, b)
	if math.Abs(got) > 1e-6 {
		t.Errorf("orthogonal vectors: want 0 got %v", got)
	}
}

func TestCosineSimilarity_OppositeVectors(t *testing.T) {
	a := []float32{1, 0}
	b := []float32{-1, 0}
	got := cosineSimilarity(a, b)
	if math.Abs(got+1.0) > 1e-6 {
		t.Errorf("opposite vectors: want -1.0 got %v", got)
	}
}

func TestCosineSimilarity_LengthMismatch(t *testing.T) {
	got := cosineSimilarity([]float32{1, 2}, []float32{1})
	if got != 0 {
		t.Errorf("length mismatch: want 0 got %v", got)
	}
}

func TestCosineSimilarity_ZeroVector(t *testing.T) {
	z := []float32{0, 0}
	a := []float32{1, 2}
	got := cosineSimilarity(z, a)
	if got != 0 {
		t.Errorf("zero vector: want 0 got %v", got)
	}
}
