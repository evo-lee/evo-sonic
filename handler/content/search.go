package content

import (
	"html"

	"github.com/evo-lee/evo-sonic/consts"
	"github.com/evo-lee/evo-sonic/handler/binding"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/dto"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/model/property"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/service/assembler"
	"github.com/evo-lee/evo-sonic/template"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type SearchHandler struct {
	PostAssembler assembler.PostAssembler
	PostService   service.PostService
	OptionService service.OptionService
	ThemeService  service.ThemeService
}

func NewSearchHandler(
	postAssembler assembler.PostAssembler,
	postService service.PostService,
	optionService service.OptionService,
	themeService service.ThemeService,
) *SearchHandler {
	return &SearchHandler{
		PostAssembler: postAssembler,
		PostService:   postService,
		OptionService: optionService,
		ThemeService:  themeService,
	}
}

func (s *SearchHandler) Search(ctx web.Context, model template.Model) (string, error) {
	return s.search(ctx, 0, model)
}

func (s *SearchHandler) PageSearch(ctx web.Context, model template.Model) (string, error) {
	page, err := util.ParamWebInt32(ctx, "page")
	if err != nil {
		return "", err
	}
	return s.search(ctx, int(page)-1, model)
}

func (s *SearchHandler) search(ctx web.Context, pageNum int, model template.Model) (string, error) {
	keyword, err := util.MustGetWebQueryString(ctx, "keyword")
	if err != nil {
		return "", err
	}
	sort := param.Sort{}
	err = ctx.BindWith(&sort, binding.CustomFormBinding)
	if err != nil {
		return "", xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if len(sort.Fields) == 0 {
		sort = s.OptionService.GetPostSort(ctx)
	}
	defaultPageSize := s.OptionService.GetIndexPageSize(ctx)
	page := param.Page{
		PageNum:  pageNum,
		PageSize: defaultPageSize,
	}
	postQuery := param.PostQuery{
		Page:     page,
		Sort:     &sort,
		Keyword:  &keyword,
		Statuses: []*consts.PostStatus{consts.PostStatusPublished.Ptr()},
	}
	posts, total, err := s.PostService.Page(ctx, postQuery)
	if err != nil {
		return "", err
	}
	postVOs, err := s.PostAssembler.ConvertToListVO(ctx, posts)
	if err != nil {
		return "", err
	}
	model["is_search"] = true
	model["keyword"] = html.EscapeString(keyword)
	model["posts"] = dto.NewPage(postVOs, total, page)
	model["meta_keywords"] = s.OptionService.GetOrByDefault(ctx, property.SeoKeywords)
	model["meta_description"] = s.OptionService.GetOrByDefault(ctx, property.SeoDescription)
	return s.ThemeService.Render(ctx, "search")
}
