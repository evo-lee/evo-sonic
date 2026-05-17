package service

import (
	"context"

	"github.com/evo-lee/evo-sonic/consts"
)

type SheetCommentService interface {
	BaseCommentService
	CountByStatus(ctx context.Context, status consts.CommentStatus) (int64, error)
}
