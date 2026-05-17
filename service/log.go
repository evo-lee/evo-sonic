package service

import (
	"context"

	"github.com/evo-lee/evo-sonic/model/dto"
	"github.com/evo-lee/evo-sonic/model/entity"
	"github.com/evo-lee/evo-sonic/model/param"
)

type LogService interface {
	PageLog(ctx context.Context, page param.Page, sort *param.Sort) ([]*entity.Log, int64, error)
	ConvertToDTO(log *entity.Log) *dto.Log
	Clear(ctx context.Context) error
}
