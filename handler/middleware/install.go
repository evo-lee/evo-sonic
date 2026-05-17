package middleware

import (
	"net/http"

	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/property"
	"github.com/evo-lee/evo-sonic/service"
)

type InstallRedirectMiddleware struct{ optionService service.OptionService }

func NewInstallRedirectMiddleware(optionService service.OptionService) *InstallRedirectMiddleware {
	return &InstallRedirectMiddleware{optionService: optionService}
}

func (i *InstallRedirectMiddleware) Handler() web.HandlerFunc {
	skipPath := map[string]struct{}{
		"/api/admin/installations":  {},
		"/api/admin/is_installed":   {},
		"/api/admin/login/precheck": {},
	}
	return func(ctx web.Context) {
		if _, ok := skipPath[ctx.Path()]; ok {
			ctx.Next()
			return
		}
		isInstall, err := i.optionService.GetOrByDefaultWithErr(ctx, property.IsInstalled, false)
		if err != nil {
			abortWithStatusJSON(ctx, http.StatusInternalServerError, "")
			return
		}
		if !isInstall.(bool) {
			ctx.Redirect(http.StatusFound, "/admin/#install")
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
