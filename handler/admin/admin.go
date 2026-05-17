package admin

import (
	"errors"

	"github.com/go-playground/validator/v10"

	"github.com/evo-lee/evo-sonic/handler/trans"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/dto"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/model/property"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type AdminHandler struct {
	OptionService       service.OptionService
	AdminService        service.AdminService
	TwoFactorMFAService service.TwoFactorTOTPMFAService
}

func NewAdminHandler(optionService service.OptionService, adminService service.AdminService, twoFactorMFA service.TwoFactorTOTPMFAService) *AdminHandler {
	return &AdminHandler{
		OptionService:       optionService,
		AdminService:        adminService,
		TwoFactorMFAService: twoFactorMFA,
	}
}

func (a *AdminHandler) IsInstalled(ctx web.Context) (interface{}, error) {
	return a.OptionService.GetOrByDefaultWithErr(ctx.RequestContext(), property.IsInstalled, false)
}

func (a *AdminHandler) AuthPreCheck(ctx web.Context) (interface{}, error) {
	var loginParam param.LoginParam
	err := ctx.BindJSON(&loginParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.BadParam.Wrapf(err, "")
	}

	user, err := a.AdminService.Authenticate(ctx.RequestContext(), loginParam)
	if err != nil {
		return nil, err
	}
	return &dto.LoginPreCheckDTO{NeedMFACode: a.TwoFactorMFAService.UseMFA(user.MfaType)}, nil
}

func (a *AdminHandler) Auth(ctx web.Context) (interface{}, error) {
	var loginParam param.LoginParam
	err := ctx.BindJSON(&loginParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.BadParam.Wrapf(err, "").WithStatus(xerr.StatusBadRequest)
	}

	return a.AdminService.Auth(ctx.RequestContext(), loginParam)
}

func (a *AdminHandler) LogOut(ctx web.Context) (interface{}, error) {
	err := a.AdminService.ClearToken(ctx.RequestContext())
	return nil, err
}

func (a *AdminHandler) SendResetCode(ctx web.Context) (interface{}, error) {
	var resetPasswordParam param.ResetPasswordParam
	err := ctx.BindJSON(&resetPasswordParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.BadParam.Wrapf(err, "").WithStatus(xerr.StatusBadRequest)
	}
	return nil, a.AdminService.SendResetPasswordCode(ctx.RequestContext(), resetPasswordParam)
}

func (a *AdminHandler) RefreshToken(ctx web.Context) (interface{}, error) {
	refreshToken := ctx.Param("refreshToken")
	if refreshToken == "" {
		return nil, xerr.BadParam.New("refreshToken参数为空").WithStatus(xerr.StatusBadRequest).
			WithMsg("refreshToken 参数不能为空")
	}
	return a.AdminService.RefreshToken(ctx.RequestContext(), refreshToken)
}

func (a *AdminHandler) GetEnvironments(ctx web.Context) (interface{}, error) {
	return a.AdminService.GetEnvironments(ctx.RequestContext()), nil
}

func (a *AdminHandler) GetLogFiles(ctx web.Context) (interface{}, error) {
	lines, err := util.MustGetWebQueryInt64(ctx, "lines")
	if err != nil {
		return nil, err
	}
	return a.AdminService.GetLogFiles(ctx.RequestContext(), lines)
}
