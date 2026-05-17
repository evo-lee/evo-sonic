package content

import (
	"github.com/evo-lee/evo-sonic/handler/content/model"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/template"
	"github.com/evo-lee/evo-sonic/util"
)

type JournalHandler struct {
	OptionService  service.OptionService
	JournalService service.JournalService
	JournalModel   *model.JournalModel
}

func NewJournalHandler(
	optionService service.OptionService,
	journalService service.JournalService,
	journalModel *model.JournalModel,
) *JournalHandler {
	return &JournalHandler{
		OptionService:  optionService,
		JournalService: journalService,
		JournalModel:   journalModel,
	}
}

func (p *JournalHandler) JournalsPage(ctx web.Context, model template.Model) (string, error) {
	page, err := util.ParamWebInt32(ctx, "page")
	if err != nil {
		return "", err
	}
	return p.JournalModel.Journals(ctx, model, int(page-1))
}

func (p *JournalHandler) Journals(ctx web.Context, model template.Model) (string, error) {
	return p.JournalModel.Journals(ctx, model, 0)
}
