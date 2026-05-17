package admin

import (
	"errors"

	"github.com/go-playground/validator/v10"

	"github.com/evo-lee/evo-sonic/handler/trans"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type InstallHandler struct {
	InstallService service.InstallService
}

func NewInstallHandler(installService service.InstallService) *InstallHandler {
	return &InstallHandler{
		InstallService: installService,
	}
}

func (i *InstallHandler) InstallBlog(ctx web.Context) (interface{}, error) {
	var installParam param.Install
	err := ctx.BindJSON(&installParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest)
	}
	err = i.InstallService.InstallBlog(ctx.RequestContext(), installParam)
	if err != nil {
		return nil, err
	}
	return "安装完成", nil
}
