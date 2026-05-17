package admin

import (
	"errors"

	"github.com/go-playground/validator/v10"

	"github.com/evo-lee/evo-sonic/consts"
	"github.com/evo-lee/evo-sonic/handler/binding"
	"github.com/evo-lee/evo-sonic/handler/trans"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/dto"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/model/property"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/service/assembler"
	"github.com/evo-lee/evo-sonic/service/impl"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type JournalCommentHandler struct {
	JournalCommentService   service.JournalCommentService
	OptionService           service.OptionService
	JournalService          service.JournalService
	JournalCommentAssembler assembler.JournalCommentAssembler
}

func NewJournalCommentHandler(journalCommentService service.JournalCommentService, optionService service.OptionService, journalService service.JournalService, journalCommentAssembler assembler.JournalCommentAssembler) *JournalCommentHandler {
	return &JournalCommentHandler{
		JournalCommentService:   journalCommentService,
		OptionService:           optionService,
		JournalService:          journalService,
		JournalCommentAssembler: journalCommentAssembler,
	}
}

func (j *JournalCommentHandler) ListJournalComment(ctx web.Context) (interface{}, error) {
	var commentQuery param.CommentQuery
	err := ctx.BindWith(&commentQuery, binding.CustomFormBinding)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	commentQuery.Sort = &param.Sort{
		Fields: []string{"createTime,desc"},
	}
	reqCtx := ctx.RequestContext()
	comments, totalCount, err := j.JournalCommentService.Page(reqCtx, commentQuery, consts.CommentTypeJournal)
	if err != nil {
		return nil, err
	}
	commentDTOs, err := j.JournalCommentAssembler.ConvertToWithJournal(reqCtx, comments)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commentDTOs, totalCount, commentQuery.Page), nil
}

func (j *JournalCommentHandler) ListJournalCommentLatest(ctx web.Context) (interface{}, error) {
	top, err := util.MustGetWebQueryInt32(ctx, "top")
	if err != nil {
		return nil, err
	}
	commentQuery := param.CommentQuery{
		Sort: &param.Sort{Fields: []string{"createTime,desc"}},
		Page: param.Page{PageNum: 0, PageSize: int(top)},
	}
	reqCtx := ctx.RequestContext()
	comments, _, err := j.JournalCommentService.Page(reqCtx, commentQuery, consts.CommentTypeSheet)
	if err != nil {
		return nil, err
	}
	return j.JournalCommentAssembler.ConvertToWithJournal(reqCtx, comments)
}

func (j *JournalCommentHandler) ListJournalCommentAsTree(ctx web.Context) (interface{}, error) {
	journalID, err := util.ParamWebInt32(ctx, "journalID")
	if err != nil {
		return nil, err
	}
	pageNum, err := util.MustGetWebQueryInt32(ctx, "page")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	pageSize, err := j.OptionService.GetOrByDefaultWithErr(reqCtx, property.CommentPageSize, property.CommentPageSize.DefaultValue)
	if err != nil {
		return nil, err
	}
	page := param.Page{PageSize: pageSize.(int), PageNum: int(pageNum)}

	allComments, err := j.JournalCommentService.GetByContentID(reqCtx, journalID, consts.CommentTypeJournal, &param.Sort{Fields: []string{"createTime,desc"}})
	if err != nil {
		return nil, err
	}

	commentVOs, totalCount, err := j.JournalCommentAssembler.PageConvertToVOs(reqCtx, allComments, page)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commentVOs, totalCount, page), nil
}

func (j *JournalCommentHandler) ListJournalCommentWithParent(ctx web.Context) (interface{}, error) {
	journalID, err := util.ParamWebInt32(ctx, "journalID")
	if err != nil {
		return nil, err
	}
	pageNum, err := util.MustGetWebQueryInt32(ctx, "page")
	if err != nil {
		return nil, err
	}

	reqCtx := ctx.RequestContext()
	pageSize, err := j.OptionService.GetOrByDefaultWithErr(reqCtx, property.CommentPageSize, property.CommentPageSize.DefaultValue)
	if err != nil {
		return nil, err
	}

	page := param.Page{PageSize: pageSize.(int), PageNum: int(pageNum)}

	comments, totalCount, err := j.JournalCommentService.Page(reqCtx, param.CommentQuery{
		ContentID: &journalID,
		Page:      page,
		Sort:      &param.Sort{Fields: []string{"createTime,desc"}},
	}, consts.CommentTypePost)
	if err != nil {
		return nil, err
	}

	commentsWithParent, err := j.JournalCommentAssembler.ConvertToWithParentVO(reqCtx, comments)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commentsWithParent, totalCount, page), nil
}

func (j *JournalCommentHandler) CreateJournalComment(ctx web.Context) (interface{}, error) {
	var commentParam *param.AdminComment
	err := ctx.BindJSON(&commentParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	reqCtx := ctx.RequestContext()
	user, err := impl.MustGetAuthorizedUser(reqCtx)
	if err != nil || user == nil {
		return nil, err
	}
	blogURL, err := j.OptionService.GetBlogBaseURL(reqCtx)
	if err != nil {
		return nil, err
	}
	commonParam := param.Comment{
		Author:            user.Username,
		Email:             user.Email,
		AuthorURL:         blogURL,
		Content:           commentParam.Content,
		PostID:            commentParam.PostID,
		ParentID:          commentParam.ParentID,
		AllowNotification: true,
		CommentType:       consts.CommentTypeJournal,
	}
	comment, err := j.JournalCommentService.CreateBy(reqCtx, &commonParam)
	if err != nil {
		return nil, err
	}
	return j.JournalCommentAssembler.ConvertToDTO(reqCtx, comment)
}

func (j *JournalCommentHandler) UpdateJournalCommentStatus(ctx web.Context) (interface{}, error) {
	commentID, err := util.ParamWebInt32(ctx, "commentID")
	if err != nil {
		return nil, err
	}
	strStatus, err := util.ParamWebString(ctx, "status")
	if err != nil {
		return nil, err
	}
	status, err := consts.CommentStatusFromString(strStatus)
	if err != nil {
		return nil, err
	}
	return j.JournalCommentService.UpdateStatus(ctx.RequestContext(), commentID, status)
}

func (j *JournalCommentHandler) UpdateJournalComment(ctx web.Context) (interface{}, error) {
	commentID, err := util.ParamWebInt32(ctx, "commentID")
	if err != nil {
		return nil, err
	}
	var commentParam *param.Comment
	err = ctx.BindJSON(&commentParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	if commentParam.AuthorURL != "" {
		err = util.Validate.Var(commentParam.AuthorURL, "url")
		if err != nil {
			return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("url is not available")
		}
	}
	reqCtx := ctx.RequestContext()
	comment, err := j.JournalCommentService.UpdateBy(reqCtx, commentID, commentParam)
	if err != nil {
		return nil, err
	}
	return j.JournalCommentAssembler.ConvertToDTO(reqCtx, comment)
}

func (j *JournalCommentHandler) UpdateJournalStatusBatch(ctx web.Context) (interface{}, error) {
	status, err := util.ParamWebInt32(ctx, "status")
	if err != nil {
		return nil, err
	}

	ids := make([]int32, 0)
	err = ctx.BindJSON(&ids)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("post ids error")
	}
	reqCtx := ctx.RequestContext()
	comments, err := j.JournalCommentService.UpdateStatusBatch(reqCtx, ids, consts.CommentStatus(status))
	if err != nil {
		return nil, err
	}
	return j.JournalCommentAssembler.ConvertToDTOList(reqCtx, comments)
}

func (j *JournalCommentHandler) DeleteJournalComment(ctx web.Context) (interface{}, error) {
	commentID, err := util.ParamWebInt32(ctx, "commentID")
	if err != nil {
		return nil, err
	}
	return nil, j.JournalCommentService.Delete(ctx.RequestContext(), commentID)
}

func (j *JournalCommentHandler) DeleteJournalCommentBatch(ctx web.Context) (interface{}, error) {
	ids := make([]int32, 0)
	err := ctx.BindJSON(&ids)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("post ids error")
	}
	return nil, j.JournalCommentService.DeleteBatch(ctx.RequestContext(), ids)
}
