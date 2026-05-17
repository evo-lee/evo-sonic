package vector

import (
	"context"
	"encoding/binary"
	"math"
	"sort"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/evo-lee/evo-sonic/model/entity"
	ai "github.com/evo-lee/evo-sonic/service/ai"
)

// DBStore persists embedding vectors as binary blobs and performs
// cosine-similarity search in Go (adequate for blog scale, no pgvector needed).
type DBStore struct {
	db *gorm.DB
}

func NewDBStore(db *gorm.DB) ai.VectorStore {
	return &DBStore{db: db}
}

func (s *DBStore) Upsert(ctx context.Context, postID int32, model string, vec []float32) error {
	row := &entity.PostEmbedding{
		PostID: postID,
		Model:  model,
		Vector: encodeVector(vec),
	}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "post_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"model", "vector", "updated_at"}),
	}).Create(row).Error
}

func (s *DBStore) Delete(ctx context.Context, postID int32) error {
	return s.db.WithContext(ctx).
		Where("post_id = ?", postID).
		Delete(&entity.PostEmbedding{}).Error
}

func (s *DBStore) Search(ctx context.Context, query []float32, topK int) ([]ai.SearchResult, error) {
	var rows []entity.PostEmbedding
	if err := s.db.WithContext(ctx).Select("post_id", "vector").Find(&rows).Error; err != nil {
		return nil, err
	}

	type scored struct {
		postID int32
		score  float64
	}
	scores := make([]scored, 0, len(rows))
	for _, row := range rows {
		vec := decodeVector(row.Vector)
		if len(vec) == 0 {
			continue
		}
		scores = append(scores, scored{postID: row.PostID, score: cosineSimilarity(query, vec)})
	}
	sort.Slice(scores, func(i, j int) bool { return scores[i].score > scores[j].score })

	if topK > len(scores) {
		topK = len(scores)
	}
	results := make([]ai.SearchResult, topK)
	for i := range results {
		results[i] = ai.SearchResult{PostID: scores[i].postID, Score: scores[i].score}
	}
	return results, nil
}

func encodeVector(vec []float32) []byte {
	b := make([]byte, len(vec)*4)
	for i, v := range vec {
		binary.LittleEndian.PutUint32(b[i*4:], math.Float32bits(v))
	}
	return b
}

func decodeVector(b []byte) []float32 {
	if len(b)%4 != 0 {
		return nil
	}
	vec := make([]float32, len(b)/4)
	for i := range vec {
		vec[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
	}
	return vec
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		ai64 := float64(a[i])
		bi64 := float64(b[i])
		dot += ai64 * bi64
		normA += ai64 * ai64
		normB += bi64 * bi64
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
