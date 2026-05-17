package content

import (
	"github.com/evo-lee/evo-sonic/handler/content/model"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/service/assembler"
	"github.com/evo-lee/evo-sonic/template"
	"github.com/evo-lee/evo-sonic/util"
)

type CategoryHandler struct {
	OptionService       service.OptionService
	PostService         service.PostService
	PostCategoryService service.PostCategoryService
	CategoryService     service.CategoryService
	PostAssembler       assembler.PostAssembler
	PostModel           *model.PostModel
	CategoryModel       *model.CategoryModel
}

func NewCategoryHandler(
	optionService service.OptionService,
	postService service.PostService,
	categoryService service.CategoryService,
	postCategoryService service.PostCategoryService,
	postAssembler assembler.PostAssembler,
	postModel *model.PostModel,
	categoryModel *model.CategoryModel,
) *CategoryHandler {
	return &CategoryHandler{
		OptionService:       optionService,
		PostService:         postService,
		PostCategoryService: postCategoryService,
		CategoryService:     categoryService,
		PostAssembler:       postAssembler,
		PostModel:           postModel,
		CategoryModel:       categoryModel,
	}
}

func (c *CategoryHandler) Categories(ctx web.Context, model template.Model) (string, error) {
	return c.CategoryModel.ListCategories(ctx, model)
}

func (c *CategoryHandler) CategoryDetail(ctx web.Context, model template.Model) (string, error) {
	slug, err := util.ParamWebString(ctx, "slug")
	if err != nil {
		return "", err
	}
	token, _ := ctx.Cookie("authentication")
	return c.CategoryModel.CategoryDetail(ctx, model, slug, 0, token)
}

func (c *CategoryHandler) CategoryDetailPage(ctx web.Context, model template.Model) (string, error) {
	slug, err := util.ParamWebString(ctx, "slug")
	if err != nil {
		return "", err
	}

	page, err := util.ParamWebInt32(ctx, "page")
	if err != nil {
		return "", err
	}
	token, _ := ctx.Cookie("authentication")
	return c.CategoryModel.CategoryDetail(ctx, model, slug, int(page-1), token)
}
