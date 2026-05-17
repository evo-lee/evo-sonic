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

type PostHandler struct {
	OptionService        service.OptionService
	PostService          service.PostService
	PostCommentService   service.PostCommentService
	PostCommentAssembler assembler.PostCommentAssembler
}

func NewPostHandler(
	optionService service.OptionService,
	postService service.PostService,
	postCommentService service.PostCommentService,
	postCommentAssembler assembler.PostCommentAssembler,
) *PostHandler {
	return &PostHandler{
		OptionService:        optionService,
		PostService:          postService,
		PostCommentService:   postCommentService,
		PostCommentAssembler: postCommentAssembler,
	}
}

func (p *PostHandler) ListTopComment(ctx web.Context) (interface{}, error) {
	postID, err := util.ParamWebInt32(ctx, "postID")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	pageSize := p.OptionService.GetOrByDefault(reqCtx, property.CommentPageSize).(int)

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
	commentQuery.ContentID = &postID
	commentQuery.Keyword = nil
	commentQuery.CommentStatus = consts.CommentStatusPublished.Ptr()
	commentQuery.PageSize = pageSize
	commentQuery.ParentID = util.Int32Ptr(0)

	comments, totalCount, err := p.PostCommentService.Page(reqCtx, commentQuery, consts.CommentTypePost)
	if err != nil {
		return nil, err
	}
	_ = p.PostCommentAssembler.ClearSensitiveField(reqCtx, comments)
	commenVOs, err := p.PostCommentAssembler.ConvertToWithHasChildren(reqCtx, comments)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commenVOs, totalCount, commentQuery.Page), nil
}

func (p *PostHandler) ListChildren(ctx web.Context) (interface{}, error) {
	postID, err := util.ParamWebInt32(ctx, "postID")
	if err != nil {
		return nil, err
	}
	parentID, err := util.ParamWebInt32(ctx, "parentID")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	children, err := p.PostCommentService.GetChildren(reqCtx, parentID, postID, consts.CommentTypePost)
	if err != nil {
		return nil, err
	}
	_ = p.PostCommentAssembler.ClearSensitiveField(reqCtx, children)
	return p.PostCommentAssembler.ConvertToDTOList(reqCtx, children)
}

func (p *PostHandler) ListCommentTree(ctx web.Context) (interface{}, error) {
	postID, err := util.ParamWebInt32(ctx, "postID")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	pageSize := p.OptionService.GetOrByDefault(reqCtx, property.CommentPageSize).(int)

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
	commentQuery.ContentID = &postID
	commentQuery.Keyword = nil
	commentQuery.CommentStatus = consts.CommentStatusPublished.Ptr()
	commentQuery.PageSize = pageSize
	commentQuery.ParentID = util.Int32Ptr(0)

	allComments, err := p.PostCommentService.GetByContentID(reqCtx, postID, consts.CommentTypePost, commentQuery.Sort)
	if err != nil {
		return nil, err
	}
	_ = p.PostCommentAssembler.ClearSensitiveField(reqCtx, allComments)
	commentVOs, total, err := p.PostCommentAssembler.PageConvertToVOs(reqCtx, allComments, commentQuery.Page)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(commentVOs, total, commentQuery.Page), nil
}

func (p *PostHandler) ListComment(ctx web.Context) (interface{}, error) {
	postID, err := util.ParamWebInt32(ctx, "postID")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	pageSize := p.OptionService.GetOrByDefault(reqCtx, property.CommentPageSize).(int)

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
	commentQuery.ContentID = &postID
	commentQuery.Keyword = nil
	commentQuery.CommentStatus = consts.CommentStatusPublished.Ptr()
	commentQuery.PageSize = pageSize
	commentQuery.ParentID = util.Int32Ptr(0)

	comments, total, err := p.PostCommentService.Page(reqCtx, commentQuery, consts.CommentTypePost)
	if err != nil {
		return nil, err
	}
	_ = p.PostCommentAssembler.ClearSensitiveField(reqCtx, comments)
	result, err := p.PostCommentAssembler.ConvertToWithParentVO(reqCtx, comments)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(result, total, commentQuery.Page), nil
}

func (p *PostHandler) CreateComment(ctx web.Context) (interface{}, error) {
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
	comment.CommentType = consts.CommentTypePost
	reqCtx := ctx.RequestContext()
	result, err := p.PostCommentService.CreateBy(reqCtx, &comment)
	if err != nil {
		return nil, err
	}
	return p.PostCommentAssembler.ConvertToDTO(reqCtx, result)
}

func (p *PostHandler) Like(ctx web.Context) (interface{}, error) {
	postID, err := util.ParamWebInt32(ctx, "postID")
	if err != nil {
		return nil, err
	}
	return nil, p.PostService.IncreaseLike(ctx.RequestContext(), postID)
}
