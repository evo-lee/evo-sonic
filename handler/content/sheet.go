package content

import (
	"github.com/evo-lee/evo-sonic/cache"
	"github.com/evo-lee/evo-sonic/handler/content/model"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/template"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type SheetHandler struct {
	OptionService service.OptionService
	SheetService  service.SheetService
	SheetModel    *model.SheetModel
	Cache         cache.Cache
}

func NewSheetHandler(
	optionService service.OptionService,
	sheetService service.SheetService,
	sheetModel *model.SheetModel,
	cache cache.Cache,
) *SheetHandler {
	return &SheetHandler{
		OptionService: optionService,
		SheetService:  sheetService,
		SheetModel:    sheetModel,
		Cache:         cache,
	}
}

func (s *SheetHandler) SheetBySlug(ctx web.Context, model template.Model) (string, error) {
	slug, err := util.ParamWebString(ctx, "slug")
	if err != nil {
		return "", err
	}
	sheet, err := s.SheetService.GetBySlug(ctx, slug)
	if err != nil {
		return "", err
	}
	token, _ := ctx.Cookie("authentication")
	return s.SheetModel.Content(ctx, sheet, token, model)
}

func (s *SheetHandler) AdminSheetBySlug(ctx web.Context, model template.Model) (string, error) {
	slug, err := util.ParamWebString(ctx, "slug")
	if err != nil {
		return "", err
	}

	token, err := util.MustGetWebQueryString(ctx, "token")
	if err != nil {
		return "", err
	}
	if token == "" {
		return "", nil
	}
	_, ok := s.Cache.Get(token)
	if !ok {
		return "", xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg("token已过期或者不存在")
	}

	sheet, err := s.SheetService.GetBySlug(ctx, slug)
	if err != nil {
		return "", err
	}

	return s.SheetModel.AdminPreviewContent(ctx, sheet, model)
}
