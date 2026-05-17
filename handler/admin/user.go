package admin

import (
	"errors"

	"github.com/go-playground/validator/v10"

	"github.com/evo-lee/evo-sonic/consts"
	"github.com/evo-lee/evo-sonic/handler/trans"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/model/vo"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/service/impl"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type UserHandler struct {
	UserService         service.UserService
	TwoFactorMFAService service.TwoFactorTOTPMFAService
}

func NewUserHandler(userService service.UserService, twoFactorMFAService service.TwoFactorTOTPMFAService) *UserHandler {
	return &UserHandler{
		UserService:         userService,
		TwoFactorMFAService: twoFactorMFAService,
	}
}

func (u *UserHandler) GetCurrentUserProfile(ctx web.Context) (interface{}, error) {
	reqCtx := ctx.RequestContext()
	user, ok := impl.GetAuthorizedUser(reqCtx)
	if !ok {
		return nil, xerr.Forbidden.New("authorized user nil").WithStatus(xerr.StatusForbidden)
	}
	return u.UserService.ConvertToDTO(reqCtx, user), nil
}

func (u *UserHandler) UpdateUserProfile(ctx web.Context) (interface{}, error) {
	userParam := &param.User{}
	err := ctx.BindJSON(userParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	reqCtx := ctx.RequestContext()
	user, err := u.UserService.Update(reqCtx, userParam)
	if err != nil {
		return nil, err
	}
	return u.UserService.ConvertToDTO(reqCtx, user), nil
}

func (u *UserHandler) UpdatePassword(ctx web.Context) (interface{}, error) {
	type Password struct {
		OldPassword string `json:"oldPassword" form:"oldPassword" binding:"gte=1,lte=100"`
		NewPassword string `json:"newPassword" form:"newPassword" binding:"gte=1,lte=100"`
	}
	passwordParam := &Password{}
	err := ctx.BindJSON(passwordParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	return nil, u.UserService.UpdatePassword(ctx.RequestContext(), passwordParam.OldPassword, passwordParam.NewPassword)
}

func (u *UserHandler) GenerateMFAQRCode(ctx web.Context) (interface{}, error) {
	type Param struct {
		MFAType *consts.MFAType `json:"mfaType"`
	}
	param := &Param{}
	err := ctx.BindJSON(param)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	if param.MFAType == nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	reqCtx := ctx.RequestContext()
	user, ok := impl.GetAuthorizedUser(reqCtx)
	if !ok || user == nil {
		return nil, xerr.Forbidden.New("").WithMsg("unauthorized").WithStatus(xerr.StatusForbidden)
	}

	mfaFactorAuthDTO := &vo.MFAFactorAuth{}
	if *param.MFAType == consts.MFATFATotp {
		key, url, err := u.TwoFactorMFAService.GenerateOTPKey(reqCtx, user.Nickname)
		if err != nil {
			return nil, err
		}
		mfaFactorAuthDTO.MFAType = consts.MFATFATotp
		mfaFactorAuthDTO.OptAuthURL = url
		mfaFactorAuthDTO.MFAKey = key
		qrCode, err := u.TwoFactorMFAService.GenerateMFAQRCode(reqCtx, url)
		if err != nil {
			return nil, err
		}
		mfaFactorAuthDTO.QRImage = qrCode
		return mfaFactorAuthDTO, nil
	} else {
		return nil, xerr.WithMsg(nil, "Not supported authentication").WithStatus(xerr.StatusBadRequest)
	}
}

func (u *UserHandler) UpdateMFA(ctx web.Context) (interface{}, error) {
	type Param struct {
		MFAType  *consts.MFAType `json:"mfaType" form:"mfaType"`
		MFAKey   string          `json:"mfaKey" form:"mfaKey"`
		AuthCode string          `json:"authcode" form:"authcode" binding:"gte=6,lte=6"`
	}
	mfaParam := &Param{}
	err := ctx.BindJSON(mfaParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	if mfaParam.MFAType == nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	return nil, u.UserService.UpdateMFA(ctx.RequestContext(), mfaParam.MFAKey, *mfaParam.MFAType, mfaParam.AuthCode)
}
