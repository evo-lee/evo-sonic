package entity

import "time"

const TableNamePostEmbedding = "post_embedding"

// PostEmbedding stores the embedding vector for a post.
// The vector is binary-encoded (little-endian float32 array) for all DB types.
type PostEmbedding struct {
	ID        int32     `gorm:"column:id;primaryKey;autoIncrement"`
	PostID    int32     `gorm:"column:post_id;uniqueIndex;not null"`
	Model     string    `gorm:"column:model;size:100;not null"`
	Vector    []byte    `gorm:"column:vector;type:blob;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (*PostEmbedding) TableName() string { return TableNamePostEmbedding }
