package admin

import (
	"errors"

	"github.com/go-playground/validator/v10"

	"github.com/evo-lee/evo-sonic/handler/trans"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/dto"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type PhotoHandler struct {
	PhotoService service.PhotoService
}

func NewPhotoHandler(photoService service.PhotoService) *PhotoHandler {
	return &PhotoHandler{
		PhotoService: photoService,
	}
}

func (p *PhotoHandler) ListPhoto(ctx web.Context) (interface{}, error) {
	sort := param.Sort{}
	err := ctx.BindQuery(&sort)
	if err != nil {
		return nil, xerr.WithMsg(err, "sort parameter error").WithStatus(xerr.StatusBadRequest)
	}
	if len(sort.Fields) == 0 {
		sort.Fields = append(sort.Fields, "createTime,desc")
	}
	reqCtx := ctx.RequestContext()
	photos, err := p.PhotoService.List(reqCtx, &sort)
	if err != nil {
		return nil, err
	}
	return p.PhotoService.ConvertToDTOs(reqCtx, photos), nil
}

func (p *PhotoHandler) PagePhotos(ctx web.Context) (interface{}, error) {
	type Param struct {
		param.Page
		param.Sort
	}
	param := Param{}
	err := ctx.BindQuery(&param)
	if err != nil {
		return nil, xerr.WithMsg(err, "parameter error").WithStatus(xerr.StatusBadRequest)
	}
	if len(param.Fields) == 0 {
		param.Fields = append(param.Fields, "createTime,desc")
	}
	reqCtx := ctx.RequestContext()
	photos, totalCount, err := p.PhotoService.Page(reqCtx, param.Page, &param.Sort)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(p.PhotoService.ConvertToDTOs(reqCtx, photos), totalCount, param.Page), nil
}

func (p *PhotoHandler) GetPhotoByID(ctx web.Context) (interface{}, error) {
	id, err := util.ParamWebInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	photo, err := p.PhotoService.GetByID(reqCtx, id)
	if err != nil {
		return nil, err
	}
	return p.PhotoService.ConvertToDTO(reqCtx, photo), nil
}

func (p *PhotoHandler) CreatePhoto(ctx web.Context) (interface{}, error) {
	photoParam := &param.Photo{}
	err := ctx.BindJSON(photoParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	reqCtx := ctx.RequestContext()
	photo, err := p.PhotoService.Create(reqCtx, photoParam)
	if err != nil {
		return nil, err
	}
	return p.PhotoService.ConvertToDTO(reqCtx, photo), nil
}

func (p *PhotoHandler) CreatePhotoBatch(ctx web.Context) (interface{}, error) {
	photosParam := make([]*param.Photo, 0)
	err := ctx.BindJSON(&photosParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	reqCtx := ctx.RequestContext()
	photos, err := p.PhotoService.CreateBatch(reqCtx, photosParam)
	if err != nil {
		return nil, err
	}
	return p.PhotoService.ConvertToDTOs(reqCtx, photos), nil
}

func (p *PhotoHandler) UpdatePhoto(ctx web.Context) (interface{}, error) {
	id, err := util.ParamWebInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	photoParam := &param.Photo{}
	err = ctx.BindJSON(photoParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	reqCtx := ctx.RequestContext()
	photo, err := p.PhotoService.Update(reqCtx, id, photoParam)
	if err != nil {
		return nil, err
	}
	return p.PhotoService.ConvertToDTO(reqCtx, photo), nil
}

func (p *PhotoHandler) DeletePhoto(ctx web.Context) (interface{}, error) {
	id, err := util.ParamWebInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	return nil, p.PhotoService.Delete(ctx.RequestContext(), id)
}

func (p *PhotoHandler) DeletePhotoBatch(ctx web.Context) (interface{}, error) {
	photosParam := make([]int32, 0)
	err := ctx.BindJSON(&photosParam)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	for _, id := range photosParam {
		err := p.PhotoService.Delete(ctx.RequestContext(), id)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (p *PhotoHandler) ListPhotoTeams(ctx web.Context) (interface{}, error) {
	return p.PhotoService.ListTeams(ctx.RequestContext())
}
