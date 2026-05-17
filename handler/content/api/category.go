package api

import (
	"github.com/evo-lee/evo-sonic/consts"
	"github.com/evo-lee/evo-sonic/handler/binding"
	"github.com/evo-lee/evo-sonic/handler/content/authentication"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/dto"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/service/assembler"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type CategoryHandler struct {
	PostService            service.PostService
	CategoryService        service.CategoryService
	CategoryAuthentication authentication.CategoryAuthentication
	PostAssembler          assembler.PostAssembler
}

func NewCategoryHandler(postService service.PostService, categoryService service.CategoryService, categoryAuthentication *authentication.CategoryAuthentication, postAssembler assembler.PostAssembler) *CategoryHandler {
	return &CategoryHandler{
		PostService:            postService,
		CategoryService:        categoryService,
		CategoryAuthentication: *categoryAuthentication,
		PostAssembler:          postAssembler,
	}
}

func (c *CategoryHandler) ListCategories(ctx web.Context) (interface{}, error) {
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
		categoryQuery.Sort = &param.Sort{Fields: []string{"updateTime,desc"}}
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

func (c *CategoryHandler) ListPosts(ctx web.Context) (interface{}, error) {
	slug, err := util.ParamWebString(ctx, "slug")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	category, err := c.CategoryService.GetBySlug(reqCtx, slug)
	if err != nil {
		return nil, err
	}
	postQuery := param.PostQuery{}
	err = ctx.BindWith(&postQuery, binding.CustomFormBinding)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if postQuery.Sort == nil {
		postQuery.Sort = &param.Sort{Fields: []string{"topPriority,desc", "updateTime,desc"}}
	}
	password, _ := util.MustGetWebQueryString(ctx, "password")

	if category.Type == consts.CategoryTypeIntimate {
		token, _ := ctx.Cookie("authentication")
		if authenticated, _ := c.CategoryAuthentication.IsAuthenticated(reqCtx, token, category.ID); !authenticated {
			token, err := c.CategoryAuthentication.Authenticate(reqCtx, token, category.ID, password)
			if err != nil {
				return nil, err
			}
			ctx.SetCookie("authentication", token, 1800, "/", "", false, true)
		}
	}
	postQuery.WithPassword = util.BoolPtr(false)
	postQuery.Statuses = []*consts.PostStatus{consts.PostStatusPublished.Ptr(), consts.PostStatusIntimate.Ptr()}
	posts, totalCount, err := c.PostService.Page(reqCtx, postQuery)
	if err != nil {
		return nil, err
	}
	postVOs, err := c.PostAssembler.ConvertToListVO(reqCtx, posts)
	return dto.NewPage(postVOs, totalCount, postQuery.Page), err
}
