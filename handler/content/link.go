package content

import (
	"github.com/evo-lee/evo-sonic/handler/content/model"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/template"
)

type LinkHandler struct {
	LinkModel *model.LinkModel
}

func NewLinkHandler(
	linkModel *model.LinkModel,
) *LinkHandler {
	return &LinkHandler{
		LinkModel: linkModel,
	}
}

func (t *LinkHandler) Link(ctx web.Context, model template.Model) (string, error) {
	return t.LinkModel.Links(ctx, model)
}
