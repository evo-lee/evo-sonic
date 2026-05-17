package api

import (
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/property"
	"github.com/evo-lee/evo-sonic/service"
)

type OptionHandler struct {
	OptionService service.OptionService
}

func NewOptionHandler(
	optionService service.OptionService,
) *OptionHandler {
	return &OptionHandler{
		OptionService: optionService,
	}
}

func (o *OptionHandler) Comment(ctx web.Context) (interface{}, error) {
	result := make(map[string]interface{})

	reqCtx := ctx.RequestContext()
	result[property.CommentGravatarSource.KeyValue] = o.OptionService.GetOrByDefault(reqCtx, property.CommentGravatarSource)
	result[property.CommentGravatarDefault.KeyValue] = o.OptionService.GetOrByDefault(reqCtx, property.CommentGravatarDefault)
	result[property.CommentContentPlaceholder.KeyValue] = o.OptionService.GetOrByDefault(reqCtx, property.CommentContentPlaceholder)
	return result, nil
}
