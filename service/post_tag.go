package service

import (
	"context"

	"github.com/evo-lee/evo-sonic/consts"
	"github.com/evo-lee/evo-sonic/model/dto"
	"github.com/evo-lee/evo-sonic/model/entity"
	"github.com/evo-lee/evo-sonic/model/param"
)

type PostTagService interface {
	PagePost(ctx context.Context, postQuery param.PostQuery) ([]*entity.Post, int64, error)
	ListTagMapByPostID(ctx context.Context, postIDs []int32) (map[int32][]*entity.Tag, error)
	ListTagByPostID(ctx context.Context, postID int32) ([]*entity.Tag, error)
	ListAllTagWithPostCount(ctx context.Context, sort *param.Sort) ([]*dto.TagWithPostCount, error)
	ListPostByTagID(ctx context.Context, tagID int32, status consts.PostStatus) ([]*entity.Post, error)
}
