package admin

import (
	"errors"

	"github.com/go-playground/validator/v10"

	"github.com/evo-lee/evo-sonic/handler/trans"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type TagHandler struct {
	PostTagService service.PostTagService
	TagService     service.TagService
}

func NewTagHandler(postTagService service.PostTagService, tagService service.TagService) *TagHandler {
	return &TagHandler{
		PostTagService: postTagService,
		TagService:     tagService,
	}
}

func (t *TagHandler) ListTags(ctx web.Context) (interface{}, error) {
	sort := param.Sort{}
	err := ctx.BindQuery(&sort)
	if err != nil {
		return nil, xerr.WithMsg(err, "sort parameter error").WithStatus(xerr.StatusBadRequest)
	}
	if len(sort.Fields) == 0 {
		sort.Fields = append(sort.Fields, "createTime,desc")
	}
	more, _ := util.MustGetWebQueryBool(ctx, "more")
	reqCtx := ctx.RequestContext()
	if more {
		return t.PostTagService.ListAllTagWithPostCount(reqCtx, &sort)
	}
	tags, err := t.TagService.ListAll(reqCtx, &sort)
	if err != nil {
		return nil, err
	}
	return t.TagService.ConvertToDTOs(reqCtx, tags)
}

func (t *TagHandler) GetTagByID(ctx web.Context) (interface{}, error) {
	id, err := util.ParamWebInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	tag, err := t.TagService.GetByID(reqCtx, id)
	if err != nil {
		return nil, err
	}
	return t.TagService.ConvertToDTO(reqCtx, tag)
}

func (t *TagHandler) CreateTag(ctx web.Context) (interface{}, error) {
	tagParam := &param.Tag{}
	err := ctx.BindJSON(tagParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	reqCtx := ctx.RequestContext()
	tag, err := t.TagService.Create(reqCtx, tagParam)
	if err != nil {
		return nil, err
	}
	return t.TagService.ConvertToDTO(reqCtx, tag)
}

func (t *TagHandler) UpdateTag(ctx web.Context) (interface{}, error) {
	id, err := util.ParamWebInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	tagParam := &param.Tag{}
	err = ctx.BindJSON(tagParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	reqCtx := ctx.RequestContext()
	tag, err := t.TagService.Update(reqCtx, id, tagParam)
	if err != nil {
		return nil, err
	}
	return t.TagService.ConvertToDTO(reqCtx, tag)
}

func (t *TagHandler) DeleteTag(ctx web.Context) (interface{}, error) {
	id, err := util.ParamWebInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	return nil, t.TagService.Delete(ctx.RequestContext(), id)
}
