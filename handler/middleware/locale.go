package middleware

import (
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/i18n"
	"github.com/evo-lee/evo-sonic/model/property"
	"github.com/evo-lee/evo-sonic/service"
)

const (
	LocaleKey    = "locale"
	LocaleHeader = "Content-Language"
)

type LocaleMiddleware struct{ optionService service.OptionService }

func NewLocaleMiddleware(optionService service.OptionService) *LocaleMiddleware {
	return &LocaleMiddleware{optionService: optionService}
}

func (m *LocaleMiddleware) apply(ctx web.Context) {
	preferred := ""
	options, err := m.optionService.ListAllOption(ctx.RequestContext())
	if err == nil {
		for _, option := range options {
			if option.Key == property.BlogLocale.KeyValue {
				if value, ok := option.Value.(string); ok {
					preferred = value
				}
				break
			}
		}
	}
	locale := i18n.ResolveLocale(preferred, ctx.Header("Accept-Language"), i18n.DefaultLocale)
	ctx.Set(LocaleKey, locale)
	ctx.SetHeader(LocaleHeader, locale)
	ctx.Next()
}

func (m *LocaleMiddleware) Handler() web.HandlerFunc { return m.apply }

func GetLocale(ctx web.Context) string {
	value, ok := ctx.Get(LocaleKey)
	if !ok {
		return i18n.DefaultLocale
	}
	if locale, ok := value.(string); ok {
		return i18n.ResolveLocale(locale, "", i18n.DefaultLocale)
	}
	return i18n.DefaultLocale
}
