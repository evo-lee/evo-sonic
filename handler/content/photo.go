package content

import (
	"github.com/evo-lee/evo-sonic/handler/content/model"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/template"
	"github.com/evo-lee/evo-sonic/util"
)

type PhotoHandler struct {
	OptionService service.OptionService
	PhotoService  service.PhotoService
	PhotoModel    *model.PhotoModel
}

func NewPhotoHandler(
	optionService service.OptionService,
	photoService service.PhotoService,
	photoModel *model.PhotoModel,
) *PhotoHandler {
	return &PhotoHandler{
		OptionService: optionService,
		PhotoService:  photoService,
		PhotoModel:    photoModel,
	}
}

func (p *PhotoHandler) PhotosPage(ctx web.Context, model template.Model) (string, error) {
	page, err := util.ParamWebInt32(ctx, "page")
	if err != nil {
		return "", err
	}
	return p.PhotoModel.Photos(ctx, model, int(page-1))
}

func (p *PhotoHandler) Phtotos(ctx web.Context, model template.Model) (string, error) {
	return p.PhotoModel.Photos(ctx, model, 0)
}
