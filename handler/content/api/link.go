package api

import (
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/service"
)

type LinkHandler struct {
	LinkService service.LinkService
}

func NewLinkHandler(linkService service.LinkService) *LinkHandler {
	return &LinkHandler{
		LinkService: linkService,
	}
}

type linkParam struct {
	*param.Sort
}

func (l *LinkHandler) ListLinks(ctx web.Context) (interface{}, error) {
	p := linkParam{}
	if err := ctx.BindQuery(&p); err != nil {
		return nil, err
	}

	if p.Sort == nil || len(p.Sort.Fields) == 0 {
		p.Sort = &param.Sort{
			Fields: []string{"createTime,desc"},
		}
	}
	reqCtx := ctx.RequestContext()
	links, err := l.LinkService.List(reqCtx, p.Sort)
	if err != nil {
		return nil, err
	}
	return l.LinkService.ConvertToDTOs(reqCtx, links), nil
}

func (l *LinkHandler) LinkTeamVO(ctx web.Context) (interface{}, error) {
	p := linkParam{}
	if err := ctx.BindQuery(&p); err != nil {
		return nil, err
	}

	if p.Sort == nil || len(p.Sort.Fields) == 0 {
		p.Sort = &param.Sort{
			Fields: []string{"createTime,desc"},
		}
	}
	reqCtx := ctx.RequestContext()
	links, err := l.LinkService.List(reqCtx, p.Sort)
	if err != nil {
		return nil, err
	}
	return l.LinkService.ConvertToLinkTeamVO(reqCtx, links), nil
}
