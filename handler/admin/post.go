package admin

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/evo-lee/evo-sonic/consts"
	"github.com/evo-lee/evo-sonic/handler/binding"
	"github.com/evo-lee/evo-sonic/handler/trans"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/dto"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/service/assembler"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type PostHandler struct {
	PostService   service.PostService
	PostAssembler assembler.PostAssembler
}

func NewPostHandler(postService service.PostService, postAssembler assembler.PostAssembler) *PostHandler {
	return &PostHandler{
		PostService:   postService,
		PostAssembler: postAssembler,
	}
}

func (p *PostHandler) ListPosts(ctx web.Context) (interface{}, error) {
	postQuery := param.PostQuery{}
	err := ctx.BindWith(&postQuery, binding.CustomFormBinding)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if postQuery.Sort == nil {
		postQuery.Sort = &param.Sort{Fields: []string{"topPriority,desc", "createTime,desc"}}
	}
	reqCtx := ctx.RequestContext()
	posts, totalCount, err := p.PostService.Page(reqCtx, postQuery)
	if err != nil {
		return nil, err
	}
	if postQuery.More == nil || *postQuery.More {
		postVOs, err := p.PostAssembler.ConvertToListVO(reqCtx, posts)
		return dto.NewPage(postVOs, totalCount, postQuery.Page), err
	}
	postDTOs := make([]*dto.Post, 0)
	for _, post := range posts {
		postDTO, err := p.PostAssembler.ConvertToSimpleDTO(reqCtx, post)
		if err != nil {
			return nil, err
		}
		postDTOs = append(postDTOs, postDTO)
	}
	return dto.NewPage(postDTOs, totalCount, postQuery.Page), nil
}

func (p *PostHandler) ListLatestPosts(ctx web.Context) (interface{}, error) {
	top, err := util.MustGetWebQueryInt32(ctx, "top")
	if err != nil {
		top = 10
	}
	postQuery := param.PostQuery{
		Page: param.Page{
			PageSize: int(top),
			PageNum:  0,
		},
		Sort: &param.Sort{
			Fields: []string{"createTime,desc"},
		},
		Keyword:    nil,
		CategoryID: nil,
		More:       util.BoolPtr(false),
	}
	reqCtx := ctx.RequestContext()
	posts, _, err := p.PostService.Page(reqCtx, postQuery)
	if err != nil {
		return nil, err
	}
	postMinimals := make([]*dto.PostMinimal, 0, len(posts))

	for _, post := range posts {
		postMinimal, err := p.PostAssembler.ConvertToMinimalDTO(reqCtx, post)
		if err != nil {
			return nil, err
		}
		postMinimals = append(postMinimals, postMinimal)
	}
	return postMinimals, nil
}

func (p *PostHandler) ListPostsByStatus(ctx web.Context) (interface{}, error) {
	var postQuery param.PostQuery
	err := ctx.BindWith(&postQuery, binding.CustomFormBinding)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if postQuery.Sort == nil {
		postQuery.Sort = &param.Sort{Fields: []string{"createTime,desc"}}
	}

	status, err := util.ParamWebInt32(ctx, "status")
	if err != nil {
		return nil, err
	}
	postQuery.Statuses = make([]*consts.PostStatus, 0)
	statusType := consts.PostStatus(status)
	postQuery.Statuses = append(postQuery.Statuses, &statusType)

	reqCtx := ctx.RequestContext()
	posts, totalCount, err := p.PostService.Page(reqCtx, postQuery)
	if err != nil {
		return nil, err
	}
	if postQuery.More == nil {
		*postQuery.More = false
	}
	if postQuery.More == nil {
		postVOs, err := p.PostAssembler.ConvertToListVO(reqCtx, posts)
		return dto.NewPage(postVOs, totalCount, postQuery.Page), err
	}

	postDTOs := make([]*dto.Post, 0)
	for _, post := range posts {
		postDTO, err := p.PostAssembler.ConvertToSimpleDTO(reqCtx, post)
		if err != nil {
			return nil, err
		}
		postDTOs = append(postDTOs, postDTO)
	}

	return dto.NewPage(postDTOs, totalCount, postQuery.Page), nil
}

func (p *PostHandler) GetByPostID(ctx web.Context) (interface{}, error) {
	postID, err := util.ParamWebInt32(ctx, "postID")
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	reqCtx := ctx.RequestContext()
	post, err := p.PostService.GetByPostID(reqCtx, postID)
	if err != nil {
		return nil, err
	}
	postDetailVO, err := p.PostAssembler.ConvertToDetailVO(reqCtx, post)
	if err != nil {
		return nil, err
	}
	return postDetailVO, nil
}

func (p *PostHandler) CreatePost(ctx web.Context) (interface{}, error) {
	var postParam param.Post
	err := ctx.BindJSON(&postParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}

	reqCtx := ctx.RequestContext()
	post, err := p.PostService.Create(reqCtx, &postParam)
	if err != nil {
		return nil, err
	}
	return p.PostAssembler.ConvertToDetailVO(reqCtx, post)
}

func (p *PostHandler) UpdatePost(ctx web.Context) (interface{}, error) {
	var postParam param.Post
	err := ctx.BindJSON(&postParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}

	postID, err := util.ParamWebInt32(ctx, "postID")
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}

	postDetailVO, err := p.PostService.Update(ctx.RequestContext(), postID, &postParam)
	if err != nil {
		return nil, err
	}
	return postDetailVO, nil
}

func (p *PostHandler) UpdatePostStatus(ctx web.Context) (interface{}, error) {
	postID, err := util.ParamWebInt32(ctx, "postID")
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	statusStr, err := util.ParamWebString(ctx, "status")
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	status, err := consts.PostStatusFromString(statusStr)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	if int32(status) < int32(consts.PostStatusPublished) || int32(status) > int32(consts.PostStatusIntimate) {
		return nil, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg("status error")
	}
	post, err := p.PostService.UpdateStatus(ctx.RequestContext(), postID, status)
	if err != nil {
		return nil, err
	}
	return p.PostAssembler.ConvertToMinimalDTO(ctx.RequestContext(), post)
}

func (p *PostHandler) UpdatePostStatusBatch(ctx web.Context) (interface{}, error) {
	statusStr, err := util.ParamWebString(ctx, "status")
	if err != nil {
		return nil, err
	}
	status, err := consts.PostStatusFromString(statusStr)
	if err != nil {
		return nil, err
	}
	if int32(status) < int32(consts.PostStatusPublished) || int32(status) > int32(consts.PostStatusIntimate) {
		return nil, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg("status error")
	}
	ids := make([]int32, 0)
	err = ctx.BindJSON(&ids)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("post ids error")
	}

	return p.PostService.UpdateStatusBatch(ctx.RequestContext(), status, ids)
}

func (p *PostHandler) UpdatePostDraft(ctx web.Context) (interface{}, error) {
	postID, err := util.ParamWebInt32(ctx, "postID")
	if err != nil {
		return nil, err
	}
	var postContentParam param.PostContent
	err = ctx.BindJSON(&postContentParam)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("content param error")
	}
	reqCtx := ctx.RequestContext()
	post, err := p.PostService.UpdateDraftContent(reqCtx, postID, postContentParam.Content, postContentParam.OriginalContent)
	if err != nil {
		return nil, err
	}
	return p.PostAssembler.ConvertToDetailDTO(reqCtx, post)
}

func (p *PostHandler) DeletePost(ctx web.Context) (interface{}, error) {
	postID, err := util.ParamWebInt32(ctx, "postID")
	if err != nil {
		return nil, err
	}
	return nil, p.PostService.Delete(ctx.RequestContext(), postID)
}

func (p *PostHandler) DeletePostBatch(ctx web.Context) (interface{}, error) {
	postIDs := make([]int32, 0)
	err := ctx.BindJSON(&postIDs)
	if err != nil {
		return nil, xerr.WithMsg(err, "postIDs error").WithStatus(xerr.StatusBadRequest)
	}
	return nil, p.PostService.DeleteBatch(ctx.RequestContext(), postIDs)
}

func (p *PostHandler) PreviewPost(ctx web.Context) {
	postID, err := util.ParamWebInt32(ctx, "postID")
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}
	previewPath, err := p.PostService.Preview(ctx.RequestContext(), postID)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.String(http.StatusOK, previewPath)
}
