package admin

import (
	"errors"

	"github.com/go-playground/validator/v10"

	"github.com/evo-lee/evo-sonic/handler/binding"
	"github.com/evo-lee/evo-sonic/handler/trans"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type LinkHandler struct {
	LinkService service.LinkService
}

func NewLinkHandler(linkService service.LinkService) *LinkHandler {
	return &LinkHandler{
		LinkService: linkService,
	}
}

func (l *LinkHandler) ListLinks(ctx web.Context) (interface{}, error) {
	sort := param.Sort{}
	err := ctx.BindWith(&sort, binding.CustomFormBinding)
	if err != nil {
		return nil, xerr.WithMsg(err, "sort parameter error").WithStatus(xerr.StatusBadRequest)
	}
	if len(sort.Fields) == 0 {
		sort.Fields = append(sort.Fields, "team,desc", "priority,asc")
	} else {
		sort.Fields = append(sort.Fields, "priority,asc")
	}
	reqCtx := ctx.RequestContext()
	links, err := l.LinkService.List(reqCtx, &sort)
	if err != nil {
		return nil, err
	}
	return l.LinkService.ConvertToDTOs(reqCtx, links), nil
}

func (l *LinkHandler) GetLinkByID(ctx web.Context) (interface{}, error) {
	id, err := util.ParamWebInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	link, err := l.LinkService.GetByID(reqCtx, id)
	if err != nil {
		return nil, err
	}
	return l.LinkService.ConvertToDTO(reqCtx, link), nil
}

func (l *LinkHandler) CreateLink(ctx web.Context) (interface{}, error) {
	linkParam := &param.Link{}
	err := ctx.BindJSON(linkParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	reqCtx := ctx.RequestContext()
	link, err := l.LinkService.Create(reqCtx, linkParam)
	if err != nil {
		return nil, err
	}
	return l.LinkService.ConvertToDTO(reqCtx, link), nil
}

func (l *LinkHandler) UpdateLink(ctx web.Context) (interface{}, error) {
	id, err := util.ParamWebInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	linkParam := &param.Link{}
	err = ctx.BindJSON(linkParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	reqCtx := ctx.RequestContext()
	link, err := l.LinkService.Update(reqCtx, id, linkParam)
	if err != nil {
		return nil, err
	}
	return l.LinkService.ConvertToDTO(reqCtx, link), nil
}

func (l *LinkHandler) DeleteLink(ctx web.Context) (interface{}, error) {
	id, err := util.ParamWebInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	return nil, l.LinkService.Delete(ctx.RequestContext(), id)
}

func (l *LinkHandler) ListLinkTeams(ctx web.Context) (interface{}, error) {
	return l.LinkService.ListTeams(ctx.RequestContext())
}
