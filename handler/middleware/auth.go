package middleware

import (
	"net/http"

	"github.com/evo-lee/evo-sonic/cache"
	"github.com/evo-lee/evo-sonic/consts"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/property"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type AuthMiddleware struct {
	OptionService       service.OptionService
	OneTimeTokenService service.OneTimeTokenService
	UserService         service.UserService
	Cache               cache.Cache
}

func NewAuthMiddleware(optionService service.OptionService, oneTimeTokenService service.OneTimeTokenService, cache cache.Cache, userService service.UserService) *AuthMiddleware {
	return &AuthMiddleware{OptionService: optionService, OneTimeTokenService: oneTimeTokenService, Cache: cache, UserService: userService}
}

func (a *AuthMiddleware) Handler() web.HandlerFunc {
	return func(ctx web.Context) {
		isInstalled, err := a.OptionService.GetOrByDefaultWithErr(ctx, property.IsInstalled, false)
		if err != nil {
			abortWithStatusJSON(ctx, http.StatusInternalServerError, "")
			return
		}
		if !isInstalled.(bool) {
			abortWithStatusJSON(ctx, http.StatusBadRequest, T(ctx, "auth.blog_not_initialized", "Blog is not initialized"))
			return
		}

		oneTimeToken, ok := ctx.Query(consts.OneTimeTokenQueryName)
		if ok {
			allowedURL, ok := a.OneTimeTokenService.Get(oneTimeToken)
			if !ok {
				abortWithStatusJSON(ctx, http.StatusBadRequest, T(ctx, "auth.one_time_token_not_exist_or_expired", "OneTimeToken is not exist or expired"))
				return
			}
			if ctx.Path() != allowedURL {
				abortWithStatusJSON(ctx, http.StatusBadRequest, T(ctx, "auth.one_time_token_uri_mismatch", "The one-time token does not correspond the request uri"))
				return
			}
			return
		}

		token := ctx.Header(consts.AdminTokenHeaderName)
		if token == "" {
			abortWithStatusJSON(ctx, http.StatusUnauthorized, T(ctx, "auth.not_logged_in", "Not logged in, please login first"))
			return
		}
		userID, ok := a.Cache.Get(cache.BuildTokenAccessKey(token))
		if !ok || userID == nil {
			abortWithStatusJSON(ctx, http.StatusUnauthorized, T(ctx, "auth.token_expired_or_not_exist", "Token has expired or does not exist"))
			return
		}

		user, err := a.UserService.GetByID(ctx, userID.(int32))
		if xerr.GetType(err) == xerr.NoRecord {
			abortWithStatusJSON(ctx, http.StatusUnauthorized, T(ctx, "auth.user_not_found", "User not found"))
			return
		}
		if err != nil {
			abortWithStatusJSON(ctx, http.StatusInternalServerError, "")
			return
		}
		ctx.Set(consts.AuthorizedUser, user)
	}
}
