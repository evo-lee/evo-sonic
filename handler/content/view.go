package content

import (
	"net/http"

	"github.com/evo-lee/evo-sonic/consts"
	"github.com/evo-lee/evo-sonic/handler/binding"
	"github.com/evo-lee/evo-sonic/handler/content/authentication"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/model/property"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/template"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type ViewHandler struct {
	OptionService          service.OptionService
	UserService            service.UserService
	CategoryService        service.CategoryService
	PostService            service.PostService
	ThemeService           service.ThemeService
	CategoryAuthentication *authentication.CategoryAuthentication
	PostAuthentication     *authentication.PostAuthentication
}

func NewViewHandler(
	optionService service.OptionService,
	userService service.UserService,
	categoryService service.CategoryService,
	postService service.PostService,
	themeService service.ThemeService,
	categoryAuthentication *authentication.CategoryAuthentication,
	postAuthentication *authentication.PostAuthentication,
) *ViewHandler {
	return &ViewHandler{
		OptionService:          optionService,
		UserService:            userService,
		CategoryService:        categoryService,
		PostService:            postService,
		ThemeService:           themeService,
		CategoryAuthentication: categoryAuthentication,
		PostAuthentication:     postAuthentication,
	}
}

func (v *ViewHandler) Admin(ctx web.Context) (interface{}, error) {
	// TODO
	return nil, nil
}

func (v *ViewHandler) Version(ctx web.Context) (interface{}, error) {
	return consts.SonicVersion, nil
}

func (v *ViewHandler) Install(ctx web.Context) {
	reqCtx := ctx.RequestContext()
	isInstall := v.OptionService.GetOrByDefault(reqCtx, property.IsInstalled).(bool)
	if isInstall {
		return
	}
	adminURLPath, _ := v.OptionService.GetAdminURLPath(reqCtx)
	ctx.Redirect(http.StatusTemporaryRedirect, adminURLPath+"/#install")
}

func (v *ViewHandler) Logo(ctx web.Context) (interface{}, error) {
	logo := v.OptionService.GetOrByDefault(ctx.RequestContext(), property.BlogLogo).(string)
	if logo != "" {
		ctx.Redirect(http.StatusTemporaryRedirect, logo)
	}
	return nil, nil
}

func (v *ViewHandler) Favicon(ctx web.Context) (interface{}, error) {
	favicon := v.OptionService.GetOrByDefault(ctx.RequestContext(), property.BlogFavicon).(string)
	if favicon != "" {
		ctx.Redirect(http.StatusTemporaryRedirect, favicon)
	}
	return nil, nil
}

func (v *ViewHandler) Authenticate(ctx web.Context, model template.Model) (string, error) {
	contentType, err := util.ParamWebString(ctx, "type")
	if err != nil {
		return v.authenticateErr(ctx, model, contentType, "", err)
	}
	slug, err := util.ParamWebString(ctx, "slug")
	if err != nil {
		return v.authenticateErr(ctx, model, contentType, slug, err)
	}

	var authenticationParam param.Authentication
	err = ctx.BindWith(&authenticationParam, binding.CustomFormBinding)
	if err != nil {
		return v.authenticateErr(ctx, model, "post", slug, err)
	}
	if authenticationParam.Password == "" {
		return v.authenticateErr(ctx, model, "post", slug, xerr.WithMsg(nil, "密码为空"))
	}

	token, _ := ctx.Cookie("authentication")

	switch contentType {
	case consts.EncryptTypeCategory.Name():
		token, err = v.authenticateCategory(ctx, slug, authenticationParam.Password, token)
	case consts.EncryptTypePost.Name():
		token, err = v.authenticatePost(ctx, slug, authenticationParam.Password, token)
	default:
		return v.authenticateErr(ctx, model, "post", slug, xerr.WithStatus(nil, xerr.StatusBadRequest))
	}
	if err != nil {
		return v.authenticateErr(ctx, model, contentType, slug, err)
	}
	ctx.SetCookie("authentication", token, 1800, "/", "", false, true)
	return "", nil
}

func (v *ViewHandler) authenticateCategory(ctx web.Context, slug, password, token string) (string, error) {
	category, err := v.CategoryService.GetBySlug(ctx, slug)
	if err != nil {
		return "", err
	}
	categoryDTO, err := v.CategoryService.ConvertToCategoryDTO(ctx, category)
	if err != nil {
		return "", err
	}

	token, err = v.CategoryAuthentication.Authenticate(ctx, token, category.ID, password)
	if err != nil {
		return "", err
	}

	ctx.Redirect(http.StatusFound, categoryDTO.FullPath)
	return token, nil
}

func (v *ViewHandler) authenticatePost(ctx web.Context, slug, password, token string) (string, error) {
	post, err := v.PostService.GetBySlug(ctx, slug)
	if err != nil {
		return "", err
	}
	fullPath, err := v.PostService.BuildFullPath(ctx, post)
	if err != nil {
		return "", err
	}
	token, err = v.PostAuthentication.Authenticate(ctx, token, post.ID, password)
	if err != nil {
		return "", err
	}

	ctx.Redirect(http.StatusFound, fullPath)
	return token, nil
}

func (v *ViewHandler) authenticateErr(ctx web.Context, model template.Model, aType string, slug string, err error) (string, error) {
	model["type"] = aType
	model["slug"] = slug
	model["errorMsg"] = xerr.GetMessage(err)
	if exist, err := v.ThemeService.TemplateExist(ctx, "post_password.tmpl"); err == nil && exist {
		return v.ThemeService.Render(ctx, "post_password")
	}
	return "common/template/post_password", nil
}
