package admin

import (
	"errors"

	"github.com/go-playground/validator/v10"

	"github.com/evo-lee/evo-sonic/consts"
	"github.com/evo-lee/evo-sonic/handler/trans"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type ThemeHandler struct {
	ThemeService  service.ThemeService
	OptionService service.OptionService
}

func NewThemeHandler(l service.ThemeService, o service.OptionService) *ThemeHandler {
	return &ThemeHandler{
		ThemeService:  l,
		OptionService: o,
	}
}

func (t *ThemeHandler) GetThemeByID(ctx web.Context) (interface{}, error) {
	themeID, err := util.ParamWebString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	return t.ThemeService.GetThemeByID(ctx.RequestContext(), themeID)
}

func (t *ThemeHandler) ListAllThemes(ctx web.Context) (interface{}, error) {
	return t.ThemeService.ListAllTheme(ctx.RequestContext())
}

func (t *ThemeHandler) ListActivatedThemeFile(ctx web.Context) (interface{}, error) {
	reqCtx := ctx.RequestContext()
	activatedThemeID, err := t.OptionService.GetActivatedThemeID(reqCtx)
	if err != nil {
		return nil, err
	}
	return t.ThemeService.ListThemeFiles(reqCtx, activatedThemeID)
}

func (t *ThemeHandler) ListThemeFileByID(ctx web.Context) (interface{}, error) {
	themeID, err := util.ParamWebString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	return t.ThemeService.ListThemeFiles(ctx.RequestContext(), themeID)
}

func (t *ThemeHandler) GetThemeFileContent(ctx web.Context) (interface{}, error) {
	path, err := util.MustGetWebQueryString(ctx, "path")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	activatedThemeID, err := t.OptionService.GetActivatedThemeID(reqCtx)
	if err != nil {
		return nil, err
	}
	return t.ThemeService.GetThemeFileContent(reqCtx, activatedThemeID, path)
}

func (t *ThemeHandler) GetThemeFileContentByID(ctx web.Context) (interface{}, error) {
	themeID, err := util.ParamWebString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	path, err := util.MustGetWebQueryString(ctx, "path")
	if err != nil {
		return nil, err
	}

	return t.ThemeService.GetThemeFileContent(ctx.RequestContext(), themeID, path)
}

func (t *ThemeHandler) UpdateThemeFile(ctx web.Context) (interface{}, error) {
	themeParam := &param.ThemeContent{}
	err := ctx.BindJSON(themeParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	reqCtx := ctx.RequestContext()
	activatedThemeID, err := t.OptionService.GetActivatedThemeID(reqCtx)
	if err != nil {
		return nil, err
	}
	return nil, t.ThemeService.UpdateThemeFile(reqCtx, activatedThemeID, themeParam.Path, themeParam.Content)
}

func (t *ThemeHandler) UpdateThemeFileByID(ctx web.Context) (interface{}, error) {
	themeID, err := util.ParamWebString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	themeParam := &param.ThemeContent{}
	err = ctx.BindJSON(themeParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	return nil, t.ThemeService.UpdateThemeFile(ctx.RequestContext(), themeID, themeParam.Path, themeParam.Content)
}

func (t *ThemeHandler) ListCustomSheetTemplate(ctx web.Context) (interface{}, error) {
	reqCtx := ctx.RequestContext()
	activatedThemeID, err := t.OptionService.GetActivatedThemeID(reqCtx)
	if err != nil {
		return nil, err
	}
	return t.ThemeService.ListCustomTemplates(reqCtx, activatedThemeID, consts.ThemeCustomSheetPrefix)
}

func (t *ThemeHandler) ListCustomPostTemplate(ctx web.Context) (interface{}, error) {
	reqCtx := ctx.RequestContext()
	activatedThemeID, err := t.OptionService.GetActivatedThemeID(reqCtx)
	if err != nil {
		return nil, err
	}
	return t.ThemeService.ListCustomTemplates(reqCtx, activatedThemeID, consts.ThemeCustomPostPrefix)
}

func (t *ThemeHandler) ActivateTheme(ctx web.Context) (interface{}, error) {
	themeID, err := util.ParamWebString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	return t.ThemeService.ActivateTheme(ctx.RequestContext(), themeID)
}

func (t *ThemeHandler) GetActivatedTheme(ctx web.Context) (interface{}, error) {
	reqCtx := ctx.RequestContext()
	activatedThemeID, err := t.OptionService.GetActivatedThemeID(reqCtx)
	if err != nil {
		return nil, err
	}
	return t.ThemeService.GetThemeByID(reqCtx, activatedThemeID)
}

func (t *ThemeHandler) GetActivatedThemeConfig(ctx web.Context) (interface{}, error) {
	reqCtx := ctx.RequestContext()
	activatedThemeID, err := t.OptionService.GetActivatedThemeID(reqCtx)
	if err != nil {
		return nil, err
	}
	return t.ThemeService.GetThemeConfig(reqCtx, activatedThemeID)
}

func (t *ThemeHandler) GetThemeConfigByID(ctx web.Context) (interface{}, error) {
	themeID, err := util.ParamWebString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	return t.ThemeService.GetThemeConfig(ctx.RequestContext(), themeID)
}

func (t *ThemeHandler) GetThemeConfigByGroup(ctx web.Context) (interface{}, error) {
	themeID, err := util.ParamWebString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	group, err := util.ParamWebString(ctx, "group")
	if err != nil {
		return nil, err
	}
	themeSettings, err := t.ThemeService.GetThemeConfig(ctx.RequestContext(), themeID)
	if err != nil {
		return nil, err
	}
	for _, setting := range themeSettings {
		if setting.Name == group {
			return setting.Items, nil
		}
	}
	return nil, nil
}

func (t *ThemeHandler) GetThemeConfigGroupNames(ctx web.Context) (interface{}, error) {
	themeID, err := util.ParamWebString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	themeSettings, err := t.ThemeService.GetThemeConfig(ctx.RequestContext(), themeID)
	if err != nil {
		return nil, err
	}
	groupNames := make([]string, len(themeSettings))
	for index, setting := range themeSettings {
		groupNames[index] = setting.Name
	}
	return groupNames, nil
}

func (t *ThemeHandler) GetActivatedThemeSettingMap(ctx web.Context) (interface{}, error) {
	reqCtx := ctx.RequestContext()
	activatedThemeID, err := t.OptionService.GetActivatedThemeID(reqCtx)
	if err != nil {
		return nil, err
	}
	return t.ThemeService.GetThemeSettingMap(reqCtx, activatedThemeID)
}

func (t *ThemeHandler) GetThemeSettingMapByID(ctx web.Context) (interface{}, error) {
	themeID, err := util.ParamWebString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	return t.ThemeService.GetThemeSettingMap(ctx.RequestContext(), themeID)
}

func (t *ThemeHandler) GetThemeSettingMapByGroupAndThemeID(ctx web.Context) (interface{}, error) {
	themeID, err := util.ParamWebString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	group, err := util.ParamWebString(ctx, "group")
	if err != nil {
		return nil, err
	}
	return t.ThemeService.GetThemeGroupSettingMap(ctx.RequestContext(), themeID, group)
}

func (t *ThemeHandler) SaveActivatedThemeSetting(ctx web.Context) (interface{}, error) {
	reqCtx := ctx.RequestContext()
	activatedThemeID, err := t.OptionService.GetActivatedThemeID(reqCtx)
	if err != nil {
		return nil, err
	}
	settings := make(map[string]interface{})
	err = ctx.BindJSON(&settings)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest)
	}
	return nil, t.ThemeService.SaveThemeSettings(reqCtx, activatedThemeID, settings)
}

func (t *ThemeHandler) SaveThemeSettingByID(ctx web.Context) (interface{}, error) {
	themeID, err := util.ParamWebString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	settings := make(map[string]interface{})
	err = ctx.BindJSON(&settings)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest)
	}
	return nil, t.ThemeService.SaveThemeSettings(ctx.RequestContext(), themeID, settings)
}

func (t *ThemeHandler) DeleteThemeByID(ctx web.Context) (interface{}, error) {
	themeID, err := util.ParamWebString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	isDeleteSetting, err := util.GetWebQueryBool(ctx, "deleteSettings", false)
	if err != nil {
		return nil, err
	}
	return nil, t.ThemeService.DeleteTheme(ctx.RequestContext(), themeID, isDeleteSetting)
}

func (t *ThemeHandler) UploadTheme(ctx web.Context) (interface{}, error) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return nil, xerr.WithMsg(err, "upload theme error").WithStatus(xerr.StatusBadRequest)
	}
	return t.ThemeService.UploadTheme(ctx.RequestContext(), fileHeader)
}

func (t *ThemeHandler) UpdateThemeByUpload(ctx web.Context) (interface{}, error) {
	themeID, err := util.ParamWebString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return nil, xerr.WithMsg(err, "upload theme error").WithStatus(xerr.StatusBadRequest)
	}
	return t.ThemeService.UpdateThemeByUpload(ctx.RequestContext(), themeID, fileHeader)
}

func (t *ThemeHandler) FetchTheme(ctx web.Context) (interface{}, error) {
	uri, _ := util.MustGetWebQueryString(ctx, "uri")
	return t.ThemeService.Fetch(ctx.RequestContext(), uri)
}

func (t *ThemeHandler) UpdateThemeByFetching(ctx web.Context) (interface{}, error) {
	themeID, err := util.ParamWebString(ctx, "themeID")
	if err != nil {
		return nil, err
	}
	uri, err := util.MustGetWebQueryString(ctx, "uri")
	if err != nil {
		return nil, err
	}
	return t.ThemeService.UpdateThemeByFetch(ctx.RequestContext(), themeID, uri)
}

func (t *ThemeHandler) ReloadTheme(ctx web.Context) (interface{}, error) {
	return nil, t.ThemeService.ReloadTheme(ctx.RequestContext())
}

func (t *ThemeHandler) TemplateExist(ctx web.Context) (interface{}, error) {
	template, err := util.MustGetWebQueryString(ctx, "template")
	if err != nil {
		return nil, err
	}
	return t.ThemeService.TemplateExist(ctx.RequestContext(), template)
}
