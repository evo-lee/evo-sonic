package api

import (
	"html/template"

	"github.com/evo-lee/evo-sonic/consts"
	"github.com/evo-lee/evo-sonic/handler/binding"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/dto"
	"github.com/evo-lee/evo-sonic/model/entity"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/model/property"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/service/assembler"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type JournalHandler struct {
	JournalService          service.JournalService
	JournalCommentService   service.JournalCommentService
	OptionService           service.ClientOptionService
	JournalCommentAssembler assembler.JournalCommentAssembler
}

func NewJournalHandler(
	journalService service.JournalService,
	journalCommentService service.JournalCommentService,
	optionService service.ClientOptionService,
	journalCommentAssembler assembler.JournalCommentAssembler,
) *JournalHandler {
	return &JournalHandler{
		JournalService:          journalService,
		JournalCommentService:   journalCommentService,
		OptionService:           optionService,
		JournalCommentAssembler: journalCommentAssembler,
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
	journalQuery.JournalType = consts.JournalTypePublic.Ptr()
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

func (j *JournalHandler) GetJournal(ctx web.Context) (interface{}, error) {
	journalID, err := util.ParamWebInt32(ctx, "journalID")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	journals, err := j.JournalService.GetByJournalIDs(reqCtx, []int32{journalID})
	if err != nil {
		return nil, err
	}
	if len(journals) == 0 {
		return nil, xerr.WithStatus(nil, xerr.StatusBadRequest)
	}
	journalDTOs, err := j.JournalService.ConvertToWithCommentDTOList(reqCtx, []*entity.Journal{journals[journalID]})
	if err != nil {
		return nil, err
	}
	return journalDTOs[0], nil
}

func (j *JournalHandler) ListTopComment(ctx web.Context) (interface{}, error) {
	journalID, err := util.ParamWebInt32(ctx, "journalID")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	pageSize := j.OptionService.GetOrByDefault(reqCtx, property.CommentPageSize).(int)

	commentQuery := param.CommentQuery{}
	err = ctx.BindWith(&commentQuery, binding.CustomFormBinding)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if commentQuery.Sort != nil && len(commentQuery.Fields) > 0 {
		commentQuery.Sort = &param.Sort{
			Fields: []string{"createTime,desc"},
		}
	}
	commentQuery.ContentID = &journalID
	commentQuery.Keyword = nil
	commentQuery.CommentStatus = consts.CommentStatusPublished.Ptr()
	commentQuery.PageSize = pageSize
	commentQuery.ParentID = util.Int32Ptr(0)

	comments, totalCount, err := j.JournalCommentService.Page(reqCtx, commentQuery, consts.CommentTypeJournal)
	if err != nil {
		return nil, err
	}
	_ = j.JournalCommentAssembler.ClearSensitiveField(reqCtx, comments)
	commenVOs, err := j.JournalCommentAssembler.ConvertToWithHasChildren(reqCtx, comments)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commenVOs, totalCount, commentQuery.Page), nil
}

func (j *JournalHandler) ListChildren(ctx web.Context) (interface{}, error) {
	journalID, err := util.ParamWebInt32(ctx, "journalID")
	if err != nil {
		return nil, err
	}
	parentID, err := util.ParamWebInt32(ctx, "parentID")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	children, err := j.JournalCommentService.GetChildren(reqCtx, parentID, journalID, consts.CommentTypeJournal)
	if err != nil {
		return nil, err
	}
	_ = j.JournalCommentAssembler.ClearSensitiveField(reqCtx, children)
	return j.JournalCommentAssembler.ConvertToDTOList(reqCtx, children)
}

func (j *JournalHandler) ListCommentTree(ctx web.Context) (interface{}, error) {
	journalID, err := util.ParamWebInt32(ctx, "journalID")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	pageSize := j.OptionService.GetOrByDefault(reqCtx, property.CommentPageSize).(int)

	commentQuery := param.CommentQuery{}
	err = ctx.BindWith(&commentQuery, binding.CustomFormBinding)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if commentQuery.Sort != nil && len(commentQuery.Fields) > 0 {
		commentQuery.Sort = &param.Sort{
			Fields: []string{"createTime,desc"},
		}
	}
	commentQuery.ContentID = &journalID
	commentQuery.Keyword = nil
	commentQuery.CommentStatus = consts.CommentStatusPublished.Ptr()
	commentQuery.PageSize = pageSize
	commentQuery.ParentID = util.Int32Ptr(0)

	allComments, err := j.JournalCommentService.GetByContentID(reqCtx, journalID, consts.CommentTypeJournal, commentQuery.Sort)
	if err != nil {
		return nil, err
	}
	_ = j.JournalCommentAssembler.ClearSensitiveField(reqCtx, allComments)
	commentVOs, total, err := j.JournalCommentAssembler.PageConvertToVOs(reqCtx, allComments, commentQuery.Page)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commentVOs, total, commentQuery.Page), nil
}

func (j *JournalHandler) ListComment(ctx web.Context) (interface{}, error) {
	journalID, err := util.ParamWebInt32(ctx, "journalID")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	pageSize := j.OptionService.GetOrByDefault(reqCtx, property.CommentPageSize).(int)

	commentQuery := param.CommentQuery{}
	err = ctx.BindWith(&commentQuery, binding.CustomFormBinding)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if commentQuery.Sort != nil && len(commentQuery.Fields) > 0 {
		commentQuery.Sort = &param.Sort{
			Fields: []string{"createTime,desc"},
		}
	}
	commentQuery.ContentID = &journalID
	commentQuery.Keyword = nil
	commentQuery.CommentStatus = consts.CommentStatusPublished.Ptr()
	commentQuery.PageSize = pageSize
	commentQuery.ParentID = util.Int32Ptr(0)

	comments, total, err := j.JournalCommentService.Page(reqCtx, commentQuery, consts.CommentTypeJournal)
	if err != nil {
		return nil, err
	}
	_ = j.JournalCommentAssembler.ClearSensitiveField(reqCtx, comments)
	result, err := j.JournalCommentAssembler.ConvertToWithParentVO(reqCtx, comments)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(result, total, commentQuery.Page), nil
}

func (j *JournalHandler) CreateComment(ctx web.Context) (interface{}, error) {
	p := param.Comment{}
	err := ctx.BindJSON(&p)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if p.AuthorURL != "" {
		err = util.Validate.Var(p.AuthorURL, "http_url")
		if err != nil {
			return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
		}
	}
	p.Author = template.HTMLEscapeString(p.Author)
	p.AuthorURL = template.HTMLEscapeString(p.AuthorURL)
	p.Content = template.HTMLEscapeString(p.Content)
	p.Email = template.HTMLEscapeString(p.Email)
	p.CommentType = consts.CommentTypeJournal
	reqCtx := ctx.RequestContext()
	result, err := j.JournalCommentService.CreateBy(reqCtx, &p)
	if err != nil {
		return nil, err
	}
	return j.JournalCommentAssembler.ConvertToDTO(reqCtx, result)
}

func (j *JournalHandler) Like(ctx web.Context) (interface{}, error) {
	journalID, err := util.ParamWebInt32(ctx, "journalID")
	if err != nil {
		return nil, err
	}
	err = j.JournalService.IncreaseLike(ctx.RequestContext(), journalID)
	if err != nil {
		return nil, err
	}
	return nil, err
}
