package content

import (
	"strconv"

	"github.com/evo-lee/evo-sonic/cache"
	"github.com/evo-lee/evo-sonic/consts"
	"github.com/evo-lee/evo-sonic/handler/content/model"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/entity"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/service/assembler"
	"github.com/evo-lee/evo-sonic/template"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type ArchiveHandler struct {
	OptionService       service.OptionService
	PostService         service.PostService
	PostCategoryService service.PostCategoryService
	CategoryService     service.CategoryService
	PostAssembler       assembler.PostAssembler
	PostModel           *model.PostModel
	Cache               cache.Cache
}

func NewArchiveHandler(
	optionService service.OptionService,
	postService service.PostService,
	categoryService service.CategoryService,
	postCategoryService service.PostCategoryService,
	postAssembler assembler.PostAssembler,
	postModel *model.PostModel,
	cache cache.Cache,
) *ArchiveHandler {
	return &ArchiveHandler{
		OptionService:       optionService,
		PostService:         postService,
		PostCategoryService: postCategoryService,
		CategoryService:     categoryService,
		PostAssembler:       postAssembler,
		PostModel:           postModel,
		Cache:               cache,
	}
}

func (a *ArchiveHandler) Archives(ctx web.Context, model template.Model) (string, error) {
	return a.PostModel.Archives(ctx, 0, model)
}

func (a *ArchiveHandler) ArchivesPage(ctx web.Context, model template.Model) (string, error) {
	page, err := util.ParamWebInt32(ctx, "page")
	if err != nil {
		return "", err
	}
	return a.PostModel.Archives(ctx, int(page-1), model)
}

func (a *ArchiveHandler) ArchivesBySlug(ctx web.Context, model template.Model) (string, error) {
	slug, err := util.ParamWebString(ctx, "slug")
	if err != nil {
		return "", err
	}

	postPermalinkType, err := a.OptionService.GetPostPermalinkType(ctx)
	if err != nil {
		return "", err
	}
	var post *entity.Post
	if postPermalinkType == consts.PostPermalinkTypeDefault {
		post, err = a.PostService.GetBySlug(ctx, slug)
		if err != nil {
			return "", err
		}
	} else if postPermalinkType == consts.PostPermalinkTypeID {
		postID, err := strconv.ParseInt(slug, 10, 32)
		if err != nil {
			return "", err
		}
		post, err = a.PostService.GetByPostID(ctx, int32(postID))
		if err != nil {
			return "", err
		}
	}
	token, _ := ctx.Cookie("authentication")
	return a.PostModel.Content(ctx, post, token, model)
}

// AdminArchivesBySlug It can only be used in the console  to preview articles
func (a *ArchiveHandler) AdminArchivesBySlug(ctx web.Context, model template.Model) (string, error) {
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
	_, ok := a.Cache.Get(token)
	if !ok {
		return "", xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg("token已过期或者不存在")
	}

	postPermalinkType, err := a.OptionService.GetPostPermalinkType(ctx)
	if err != nil {
		return "", err
	}
	var post *entity.Post
	if postPermalinkType == consts.PostPermalinkTypeDefault {
		post, err = a.PostService.GetBySlug(ctx, slug)
		if err != nil {
			return "", err
		}
	} else if postPermalinkType == consts.PostPermalinkTypeID {
		postID, err := strconv.ParseInt(slug, 10, 32)
		if err != nil {
			return "", err
		}
		post, err = a.PostService.GetByPostID(ctx, int32(postID))
		if err != nil {
			return "", err
		}
	}
	return a.PostModel.AdminPreview(ctx, post, model)
}
