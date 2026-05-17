package api

import (
	"github.com/evo-lee/evo-sonic/consts"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/service/assembler"
)

type ArchiveHandler struct {
	PostService   service.PostService
	PostAssembler assembler.PostAssembler
}

func NewArchiveHandler(postService service.PostService, postAssemeber assembler.PostAssembler) *ArchiveHandler {
	return &ArchiveHandler{
		PostService:   postService,
		PostAssembler: postAssemeber,
	}
}

func (a *ArchiveHandler) ListYearArchives(ctx web.Context) (interface{}, error) {
	posts, err := a.PostService.GetByStatus(ctx.RequestContext(), []consts.PostStatus{consts.PostStatusPublished}, consts.PostTypePost, nil)
	if err != nil {
		return nil, err
	}
	return a.PostAssembler.ConvertToArchiveYearVOs(ctx.RequestContext(), posts)
}

func (a *ArchiveHandler) ListMonthArchives(ctx web.Context) (interface{}, error) {
	posts, err := a.PostService.GetByStatus(ctx.RequestContext(), []consts.PostStatus{consts.PostStatusPublished}, consts.PostTypePost, nil)
	if err != nil {
		return nil, err
	}
	return a.PostAssembler.ConvertTOArchiveMonthVOs(ctx.RequestContext(), posts)
}
