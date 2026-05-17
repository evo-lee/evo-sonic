package service

import (
	"context"

	"github.com/evo-lee/evo-sonic/model/param"
)

type InstallService interface {
	InstallBlog(ctx context.Context, installParam param.Install) error
}
