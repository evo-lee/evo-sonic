package admin

import (
	"errors"

	"github.com/go-playground/validator/v10"

	"github.com/evo-lee/evo-sonic/handler/trans"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type CategoryHandler struct {
	CategoryService service.CategoryService
}

func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		CategoryService: categoryService,
	}
}

func (c *CategoryHandler) GetCategoryByID(ctx web.Context) (interface{}, error) {
	id, err := util.ParamWebInt32(ctx, "categoryID")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	category, err := c.CategoryService.GetByID(reqCtx, id)
	if err != nil {
		return nil, err
	}
	return c.CategoryService.ConvertToCategoryDTO(reqCtx, category)
}

func (c *CategoryHandler) ListAllCategory(ctx web.Context) (interface{}, error) {
	categoryQuery := struct {
		*param.Sort
		More *bool `json:"more" form:"more"`
	}{}

	err := ctx.BindQuery(&categoryQuery)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	reqCtx := ctx.RequestContext()
	if categoryQuery.Sort == nil || len(categoryQuery.Sort.Fields) == 0 {
		categoryQuery.Sort = &param.Sort{Fields: []string{"priority,asc"}}
	}
	if categoryQuery.More != nil && *categoryQuery.More {
		return c.CategoryService.ListCategoryWithPostCountDTO(reqCtx, categoryQuery.Sort)
	}
	categories, err := c.CategoryService.ListAll(reqCtx, categoryQuery.Sort)
	if err != nil {
		return nil, err
	}
	return c.CategoryService.ConvertToCategoryDTOs(reqCtx, categories)
}

func (c *CategoryHandler) ListAsTree(ctx web.Context) (interface{}, error) {
	var sort param.Sort
	err := ctx.BindQuery(&sort)
	if err != nil {
		return nil, err
	}
	if len(sort.Fields) == 0 {
		sort.Fields = append(sort.Fields, "priority,asc")
	}
	return c.CategoryService.ListAsTree(ctx.RequestContext(), &sort, false)
}

func (c *CategoryHandler) CreateCategory(ctx web.Context) (interface{}, error) {
	var categoryParam param.Category
	err := ctx.BindJSON(&categoryParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest)
	}
	reqCtx := ctx.RequestContext()
	category, err := c.CategoryService.Create(reqCtx, &categoryParam)
	if err != nil {
		return nil, err
	}
	return c.CategoryService.ConvertToCategoryDTO(reqCtx, category)
}

func (c *CategoryHandler) UpdateCategory(ctx web.Context) (interface{}, error) {
	var categoryParam param.Category
	err := ctx.BindJSON(&categoryParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest)
	}
	categoryID, err := util.ParamWebInt32(ctx, "categoryID")
	if err != nil {
		return nil, err
	}
	categoryParam.ID = categoryID
	reqCtx := ctx.RequestContext()
	category, err := c.CategoryService.Update(reqCtx, &categoryParam)
	if err != nil {
		return nil, err
	}
	return c.CategoryService.ConvertToCategoryDTO(reqCtx, category)
}

func (c *CategoryHandler) UpdateCategoryBatch(ctx web.Context) (interface{}, error) {
	categoryParams := make([]*param.Category, 0)
	err := ctx.BindJSON(&categoryParams)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	reqCtx := ctx.RequestContext()
	categories, err := c.CategoryService.UpdateBatch(reqCtx, categoryParams)
	if err != nil {
		return nil, err
	}
	return c.CategoryService.ConvertToCategoryDTOs(reqCtx, categories)
}

func (c *CategoryHandler) DeleteCategory(ctx web.Context) (interface{}, error) {
	categoryID, err := util.ParamWebInt32(ctx, "categoryID")
	if err != nil {
		return nil, err
	}
	return nil, c.CategoryService.Delete(ctx.RequestContext(), categoryID)
}
