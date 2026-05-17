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

type PostCommentHandler struct {
	PostCommentService   service.PostCommentService
	OptionService        service.OptionService
	PostService          service.PostService
	PostAssembler        assembler.PostAssembler
	PostCommentAssembler assembler.PostCommentAssembler
}

func NewPostCommentHandler(
	postCommentHandler service.PostCommentService,
	optionService service.OptionService,
	postService service.PostService,
	postAssembler assembler.PostAssembler,
	postCommentAssembler assembler.PostCommentAssembler,
) *PostCommentHandler {
	return &PostCommentHandler{
		PostCommentService:   postCommentHandler,
		OptionService:        optionService,
		PostService:          postService,
		PostAssembler:        postAssembler,
		PostCommentAssembler: postCommentAssembler,
	}
}

func (p *PostCommentHandler) ListPostComment(ctx web.Context) (interface{}, error) {
	var commentQuery param.CommentQuery
	err := ctx.BindWith(&commentQuery, binding.CustomFormBinding)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	commentQuery.Sort = &param.Sort{
		Fields: []string{"createTime,desc"},
	}
	reqCtx := ctx.RequestContext()
	comments, totalCount, err := p.PostCommentService.Page(reqCtx, commentQuery, consts.CommentTypePost)
	if err != nil {
		return nil, err
	}
	commentDTOs, err := p.PostCommentAssembler.ConvertToWithPost(reqCtx, comments)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commentDTOs, totalCount, commentQuery.Page), nil
}

func (p *PostCommentHandler) ListPostCommentLatest(ctx web.Context) (interface{}, error) {
	top, err := util.MustGetWebQueryInt32(ctx, "top")
	if err != nil {
		return nil, err
	}
	commentQuery := param.CommentQuery{
		Sort: &param.Sort{Fields: []string{"createTime,desc"}},
		Page: param.Page{PageNum: 0, PageSize: int(top)},
	}
	reqCtx := ctx.RequestContext()
	comments, _, err := p.PostCommentService.Page(reqCtx, commentQuery, consts.CommentTypePost)
	if err != nil {
		return nil, err
	}
	return p.PostCommentAssembler.ConvertToWithPost(reqCtx, comments)
}

func (p *PostCommentHandler) ListPostCommentAsTree(ctx web.Context) (interface{}, error) {
	postID, err := util.ParamWebInt32(ctx, "postID")
	if err != nil {
		return nil, err
	}
	pageNum, err := util.MustGetWebQueryInt32(ctx, "page")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	pageSize, err := p.OptionService.GetOrByDefaultWithErr(reqCtx, property.CommentPageSize, property.CommentPageSize.DefaultValue)
	if err != nil {
		return nil, err
	}
	page := param.Page{PageSize: pageSize.(int), PageNum: int(pageNum)}
	allComments, err := p.PostCommentService.GetByContentID(reqCtx, postID, consts.CommentTypePost, &param.Sort{Fields: []string{"createTime,desc"}})
	if err != nil {
		return nil, err
	}
	commentVOs, totalCount, err := p.PostCommentAssembler.PageConvertToVOs(reqCtx, allComments, page)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commentVOs, totalCount, page), nil
}

func (p *PostCommentHandler) ListPostCommentWithParent(ctx web.Context) (interface{}, error) {
	postID, err := util.ParamWebInt32(ctx, "postID")
	if err != nil {
		return nil, err
	}
	pageNum, err := util.MustGetWebQueryInt32(ctx, "page")
	if err != nil {
		return nil, err
	}

	reqCtx := ctx.RequestContext()
	pageSize, err := p.OptionService.GetOrByDefaultWithErr(reqCtx, property.CommentPageSize, property.CommentPageSize.DefaultValue)
	if err != nil {
		return nil, err
	}

	page := param.Page{PageNum: int(pageNum), PageSize: pageSize.(int)}

	comments, totalCount, err := p.PostCommentService.Page(reqCtx, param.CommentQuery{
		ContentID: &postID,
		Page:      page,
		Sort:      &param.Sort{Fields: []string{"createTime,desc"}},
	}, consts.CommentTypePost)
	if err != nil {
		return nil, err
	}

	commentsWithParent, err := p.PostCommentAssembler.ConvertToWithParentVO(reqCtx, comments)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commentsWithParent, totalCount, page), nil
}

func (p *PostCommentHandler) CreatePostComment(ctx web.Context) (interface{}, error) {
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
	blogURL, err := p.OptionService.GetBlogBaseURL(reqCtx)
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
		CommentType:       consts.CommentTypePost,
	}
	comment, err := p.PostCommentService.CreateBy(reqCtx, &commonParam)
	if err != nil {
		return nil, err
	}
	return p.PostCommentAssembler.ConvertToDTO(reqCtx, comment)
}

func (p *PostCommentHandler) UpdatePostComment(ctx web.Context) (interface{}, error) {
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
	comment, err := p.PostCommentService.UpdateBy(reqCtx, commentID, commentParam)
	if err != nil {
		return nil, err
	}

	return p.PostCommentAssembler.ConvertToDTO(reqCtx, comment)
}

func (p *PostCommentHandler) UpdatePostCommentStatus(ctx web.Context) (interface{}, error) {
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
	return p.PostCommentService.UpdateStatus(ctx.RequestContext(), commentID, status)
}

func (p *PostCommentHandler) UpdatePostCommentStatusBatch(ctx web.Context) (interface{}, error) {
	strStatus, err := util.ParamWebString(ctx, "status")
	if err != nil {
		return nil, err
	}
	status, err := consts.CommentStatusFromString(strStatus)
	if err != nil {
		return nil, err
	}

	ids := make([]int32, 0)
	err = ctx.BindJSON(&ids)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("post ids error")
	}
	reqCtx := ctx.RequestContext()
	comments, err := p.PostCommentService.UpdateStatusBatch(reqCtx, ids, status)
	if err != nil {
		return nil, err
	}
	return p.PostCommentAssembler.ConvertToDTOList(reqCtx, comments)
}

func (p *PostCommentHandler) DeletePostComment(ctx web.Context) (interface{}, error) {
	commentID, err := util.ParamWebInt32(ctx, "commentID")
	if err != nil {
		return nil, err
	}
	return nil, p.PostCommentService.Delete(ctx.RequestContext(), commentID)
}

func (p *PostCommentHandler) DeletePostCommentBatch(ctx web.Context) (interface{}, error) {
	ids := make([]int32, 0)
	err := ctx.BindJSON(&ids)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("post ids error")
	}
	return nil, p.PostCommentService.DeleteBatch(ctx.RequestContext(), ids)
}
