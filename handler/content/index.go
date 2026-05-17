package content

import (
	"github.com/evo-lee/evo-sonic/handler/content/model"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/template"
	"github.com/evo-lee/evo-sonic/util"
)

type IndexHandler struct {
	PostModel *model.PostModel
}

func NewIndexHandler(postModel *model.PostModel) *IndexHandler {
	return &IndexHandler{
		PostModel: postModel,
	}
}

func (h *IndexHandler) Index(ctx web.Context, model template.Model) (string, error) {
	return h.PostModel.List(ctx, 0, model)
}

func (h *IndexHandler) IndexPage(ctx web.Context, model template.Model) (string, error) {
	page, err := util.ParamWebInt32(ctx, "page")
	if err != nil {
		return "", err
	}
	return h.PostModel.List(ctx, int(page)-1, model)
}
