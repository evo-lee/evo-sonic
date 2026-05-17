package admin

import (
	"errors"

	"github.com/go-playground/validator/v10"

	"github.com/evo-lee/evo-sonic/handler/binding"
	"github.com/evo-lee/evo-sonic/handler/trans"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/dto"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type JournalHandler struct {
	JournalService service.JournalService
}

func NewJournalHandler(journalService service.JournalService) *JournalHandler {
	return &JournalHandler{
		JournalService: journalService,
	}
}

func (j *JournalHandler) ListJournal(ctx web.Context) (interface{}, error) {
	var journalQuery param.JournalQuery
	err := ctx.BindWith(&journalQuery, binding.CustomFormBinding)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	journalQuery.Sort = &param.Sort{
		Fields: []string{"createTime,desc"},
	}
	reqCtx := ctx.RequestContext()
	journals, totalCount, err := j.JournalService.ListJournal(reqCtx, journalQuery)
	if err != nil {
		return nil, err
	}
	journalDTOs, err := j.JournalService.ConvertToWithCommentDTOList(reqCtx, journals)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(journalDTOs, totalCount, journalQuery.Page), nil
}

func (j *JournalHandler) ListLatestJournal(ctx web.Context) (interface{}, error) {
	top, err := util.MustGetWebQueryInt(ctx, "top")
	if err != nil {
		top = 10
	}
	journalQuery := param.JournalQuery{
		Sort: &param.Sort{Fields: []string{"createTime,desc"}},
		Page: param.Page{PageNum: 0, PageSize: top},
	}
	reqCtx := ctx.RequestContext()
	journals, _, err := j.JournalService.ListJournal(reqCtx, journalQuery)
	if err != nil {
		return nil, err
	}
	return j.JournalService.ConvertToWithCommentDTOList(reqCtx, journals)
}

func (j *JournalHandler) CreateJournal(ctx web.Context) (interface{}, error) {
	var journalParam param.Journal
	err := ctx.BindJSON(&journalParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	if journalParam.Content == "" {
		journalParam.Content = journalParam.SourceContent
	}
	journal, err := j.JournalService.Create(ctx.RequestContext(), &journalParam)
	if err != nil {
		return nil, err
	}
	return j.JournalService.ConvertToDTO(journal), nil
}

func (j *JournalHandler) UpdateJournal(ctx web.Context) (interface{}, error) {
	var journalParam param.Journal
	err := ctx.BindJSON(&journalParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}

	journalID, err := util.ParamWebInt32(ctx, "journalID")
	if err != nil {
		return nil, err
	}
	return j.JournalService.Update(ctx.RequestContext(), journalID, &journalParam)
}

func (j *JournalHandler) DeleteJournal(ctx web.Context) (interface{}, error) {
	journalID, err := util.ParamWebInt32(ctx, "journalID")
	if err != nil {
		return nil, err
	}
	return nil, j.JournalService.Delete(ctx.RequestContext(), journalID)
}
