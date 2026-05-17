package admin

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"

	"github.com/evo-lee/evo-sonic/consts"
	"github.com/evo-lee/evo-sonic/handler/binding"
	"github.com/evo-lee/evo-sonic/handler/trans"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/dto"
	"github.com/evo-lee/evo-sonic/model/entity"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/model/property"
	"github.com/evo-lee/evo-sonic/model/vo"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/service/assembler"
	"github.com/evo-lee/evo-sonic/service/impl"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type SheetCommentHandler struct {
	SheetCommentService   service.SheetCommentService
	BaseCommentService    service.BaseCommentService
	OptionService         service.OptionService
	SheetService          service.SheetService
	SheetAssembler        assembler.SheetAssembler
	SheetCommentAssembler assembler.SheetCommentAssembler
}

func NewSheetCommentHandler(
	sheetCommentService service.SheetCommentService,
	baseCommentService service.BaseCommentService,
	optionService service.OptionService,
	sheetService service.SheetService,
	sheetAssembler assembler.SheetAssembler,
	sheetCommentAssembler assembler.SheetCommentAssembler,
) *SheetCommentHandler {
	return &SheetCommentHandler{
		SheetCommentService:   sheetCommentService,
		BaseCommentService:    baseCommentService,
		OptionService:         optionService,
		SheetService:          sheetService,
		SheetAssembler:        sheetAssembler,
		SheetCommentAssembler: sheetCommentAssembler,
	}
}

func (s *SheetCommentHandler) ListSheetComment(ctx web.Context) (interface{}, error) {
	var commentQuery param.CommentQuery
	err := ctx.BindWith(&commentQuery, binding.CustomFormBinding)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	commentQuery.Sort = &param.Sort{
		Fields: []string{"createTime,desc"},
	}
	reqCtx := ctx.RequestContext()
	comments, totalCount, err := s.SheetCommentService.Page(reqCtx, commentQuery, consts.CommentTypeSheet)
	if err != nil {
		return nil, err
	}
	commentDTOs, err := s.ConvertToWithSheet(reqCtx, comments)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commentDTOs, totalCount, commentQuery.Page), nil
}

func (s *SheetCommentHandler) ListSheetCommentLatest(ctx web.Context) (interface{}, error) {
	top, err := util.MustGetWebQueryInt32(ctx, "top")
	if err != nil {
		return nil, err
	}
	commentQuery := param.CommentQuery{
		Sort: &param.Sort{Fields: []string{"createTime,desc"}},
		Page: param.Page{PageNum: 0, PageSize: int(top)},
	}
	reqCtx := ctx.RequestContext()
	comments, _, err := s.SheetCommentService.Page(reqCtx, commentQuery, consts.CommentTypeSheet)
	if err != nil {
		return nil, err
	}
	return s.ConvertToWithSheet(reqCtx, comments)
}

func (s *SheetCommentHandler) ListSheetCommentAsTree(ctx web.Context) (interface{}, error) {
	postID, err := util.ParamWebInt32(ctx, "sheetID")
	if err != nil {
		return nil, err
	}
	pageNum, err := util.MustGetWebQueryInt32(ctx, "page")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	pageSize, err := s.OptionService.GetOrByDefaultWithErr(reqCtx, property.CommentPageSize, property.CommentPageSize.DefaultValue)
	if err != nil {
		return nil, err
	}
	page := param.Page{PageSize: pageSize.(int), PageNum: int(pageNum)}

	allComments, err := s.SheetCommentService.GetByContentID(reqCtx, postID, consts.CommentTypeSheet, &param.Sort{Fields: []string{"createTime,desc"}})
	if err != nil {
		return nil, err
	}
	commentVOs, totalCount, err := s.SheetCommentAssembler.PageConvertToVOs(reqCtx, allComments, page)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commentVOs, totalCount, page), nil
}

func (s *SheetCommentHandler) ListSheetCommentWithParent(ctx web.Context) (interface{}, error) {
	postID, err := util.ParamWebInt32(ctx, "sheetID")
	if err != nil {
		return nil, err
	}
	pageNum, err := util.MustGetWebQueryInt32(ctx, "page")
	if err != nil {
		return nil, err
	}

	reqCtx := ctx.RequestContext()
	pageSize, err := s.OptionService.GetOrByDefaultWithErr(reqCtx, property.CommentPageSize, property.CommentPageSize.DefaultValue)
	if err != nil {
		return nil, err
	}
	page := param.Page{PageSize: pageSize.(int), PageNum: int(pageNum)}

	comments, totalCount, err := s.SheetCommentService.Page(reqCtx, param.CommentQuery{
		ContentID: &postID,
		Page:      page,
		Sort:      &param.Sort{Fields: []string{"createTime,desc"}},
	}, consts.CommentTypePost)
	if err != nil {
		return nil, err
	}

	commentsWithParent, err := s.SheetCommentAssembler.ConvertToWithParentVO(reqCtx, comments)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commentsWithParent, totalCount, page), nil
}

func (s *SheetCommentHandler) CreateSheetComment(ctx web.Context) (interface{}, error) {
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
	blogURL, err := s.OptionService.GetBlogBaseURL(reqCtx)
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
		CommentType:       consts.CommentTypeSheet,
	}
	comment, err := s.BaseCommentService.CreateBy(reqCtx, &commonParam)
	if err != nil {
		return nil, err
	}
	return s.SheetCommentAssembler.ConvertToDTO(reqCtx, comment)
}

func (s *SheetCommentHandler) UpdateSheetCommentStatus(ctx web.Context) (interface{}, error) {
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
	return s.SheetCommentService.UpdateStatus(ctx.RequestContext(), commentID, status)
}

func (s *SheetCommentHandler) UpdateSheetCommentStatusBatch(ctx web.Context) (interface{}, error) {
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
	comments, err := s.SheetCommentService.UpdateStatusBatch(reqCtx, ids, consts.CommentStatus(status))
	if err != nil {
		return nil, err
	}
	return s.SheetCommentAssembler.ConvertToDTOList(reqCtx, comments)
}

func (s *SheetCommentHandler) DeleteSheetComment(ctx web.Context) (interface{}, error) {
	commentID, err := util.ParamWebInt32(ctx, "commentID")
	if err != nil {
		return nil, err
	}
	return nil, s.SheetCommentService.Delete(ctx.RequestContext(), commentID)
}

func (s *SheetCommentHandler) DeleteSheetCommentBatch(ctx web.Context) (interface{}, error) {
	ids := make([]int32, 0)
	err := ctx.BindJSON(&ids)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("post ids error")
	}
	return nil, s.SheetCommentService.DeleteBatch(ctx.RequestContext(), ids)
}

func (s *SheetCommentHandler) ConvertToWithSheet(ctx context.Context, comments []*entity.Comment) ([]*vo.SheetCommentWithSheet, error) {
	postIDs := make([]int32, 0, len(comments))
	for _, comment := range comments {
		postIDs = append(postIDs, comment.PostID)
	}
	posts, err := s.SheetService.GetByPostIDs(ctx, postIDs)
	if err != nil {
		return nil, err
	}
	result := make([]*vo.SheetCommentWithSheet, 0, len(comments))
	for _, comment := range comments {
		commentDTO, err := s.SheetCommentAssembler.ConvertToDTO(ctx, comment)
		if err != nil {
			return nil, err
		}
		commentWithSheet := &vo.SheetCommentWithSheet{
			Comment: *commentDTO,
		}
		result = append(result, commentWithSheet)
		post, ok := posts[comment.PostID]
		if ok {
			commentWithSheet.PostMinimal, err = s.SheetAssembler.ConvertToMinimalDTO(ctx, post)
			if err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}
