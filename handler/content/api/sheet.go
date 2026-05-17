package api

import (
	"html/template"

	"github.com/evo-lee/evo-sonic/consts"
	"github.com/evo-lee/evo-sonic/handler/binding"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/dto"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/model/property"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/service/assembler"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type SheetHandler struct {
	OptionService         service.OptionService
	SheetService          service.SheetService
	SheetCommentService   service.SheetCommentService
	SheetCommentAssembler assembler.SheetCommentAssembler
}

func NewSheetHandler(
	optionService service.OptionService,
	sheetService service.SheetService,
	sheetCommentService service.SheetCommentService,
	sheetCommentAssembler assembler.SheetCommentAssembler,
) *SheetHandler {
	return &SheetHandler{
		OptionService:         optionService,
		SheetService:          sheetService,
		SheetCommentService:   sheetCommentService,
		SheetCommentAssembler: sheetCommentAssembler,
	}
}

func (s *SheetHandler) ListTopComment(ctx web.Context) (interface{}, error) {
	sheetID, err := util.ParamWebInt32(ctx, "sheetID")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	pageSize := s.OptionService.GetOrByDefault(reqCtx, property.CommentPageSize).(int)

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
	commentQuery.ContentID = &sheetID
	commentQuery.Keyword = nil
	commentQuery.CommentStatus = consts.CommentStatusPublished.Ptr()
	commentQuery.PageSize = pageSize
	commentQuery.ParentID = util.Int32Ptr(0)

	comments, totalCount, err := s.SheetCommentService.Page(reqCtx, commentQuery, consts.CommentTypeSheet)
	if err != nil {
		return nil, err
	}
	_ = s.SheetCommentAssembler.ClearSensitiveField(reqCtx, comments)
	commenVOs, err := s.SheetCommentAssembler.ConvertToWithHasChildren(reqCtx, comments)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commenVOs, totalCount, commentQuery.Page), nil
}

func (s *SheetHandler) ListChildren(ctx web.Context) (interface{}, error) {
	sheetID, err := util.ParamWebInt32(ctx, "sheetID")
	if err != nil {
		return nil, err
	}
	parentID, err := util.ParamWebInt32(ctx, "parentID")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	children, err := s.SheetCommentService.GetChildren(reqCtx, parentID, sheetID, consts.CommentTypeSheet)
	if err != nil {
		return nil, err
	}
	_ = s.SheetCommentAssembler.ClearSensitiveField(reqCtx, children)
	return s.SheetCommentAssembler.ConvertToDTOList(reqCtx, children)
}

func (s *SheetHandler) ListCommentTree(ctx web.Context) (interface{}, error) {
	sheetID, err := util.ParamWebInt32(ctx, "sheetID")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	pageSize := s.OptionService.GetOrByDefault(reqCtx, property.CommentPageSize).(int)

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
	commentQuery.ContentID = &sheetID
	commentQuery.Keyword = nil
	commentQuery.CommentStatus = consts.CommentStatusPublished.Ptr()
	commentQuery.PageSize = pageSize
	commentQuery.ParentID = util.Int32Ptr(0)

	allComments, err := s.SheetCommentService.GetByContentID(reqCtx, sheetID, consts.CommentTypeSheet, commentQuery.Sort)
	if err != nil {
		return nil, err
	}
	_ = s.SheetCommentAssembler.ClearSensitiveField(reqCtx, allComments)
	commentVOs, total, err := s.SheetCommentAssembler.PageConvertToVOs(reqCtx, allComments, commentQuery.Page)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commentVOs, total, commentQuery.Page), nil
}

func (s *SheetHandler) ListComment(ctx web.Context) (interface{}, error) {
	sheetID, err := util.ParamWebInt32(ctx, "sheetID")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	pageSize := s.OptionService.GetOrByDefault(reqCtx, property.CommentPageSize).(int)

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
	commentQuery.ContentID = &sheetID
	commentQuery.Keyword = nil
	commentQuery.CommentStatus = consts.CommentStatusPublished.Ptr()
	commentQuery.PageSize = pageSize
	commentQuery.ParentID = util.Int32Ptr(0)

	comments, total, err := s.SheetCommentService.Page(reqCtx, commentQuery, consts.CommentTypeSheet)
	if err != nil {
		return nil, err
	}
	_ = s.SheetCommentAssembler.ClearSensitiveField(reqCtx, comments)
	result, err := s.SheetCommentAssembler.ConvertToWithParentVO(reqCtx, comments)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(result, total, commentQuery.Page), nil
}

func (s *SheetHandler) CreateComment(ctx web.Context) (interface{}, error) {
	comment := param.Comment{}
	err := ctx.BindJSON(&comment)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if comment.AuthorURL != "" {
		err = util.Validate.Var(comment.AuthorURL, "http_url")
		if err != nil {
			return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
		}
	}
	comment.Author = template.HTMLEscapeString(comment.Author)
	comment.AuthorURL = template.HTMLEscapeString(comment.AuthorURL)
	comment.Content = template.HTMLEscapeString(comment.Content)
	comment.Email = template.HTMLEscapeString(comment.Email)
	comment.CommentType = consts.CommentTypeSheet
	reqCtx := ctx.RequestContext()
	result, err := s.SheetCommentService.CreateBy(reqCtx, &comment)
	if err != nil {
		return nil, err
	}
	return s.SheetCommentAssembler.ConvertToDTO(reqCtx, result)
}
