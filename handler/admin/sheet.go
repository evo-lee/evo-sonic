package admin

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/evo-lee/evo-sonic/consts"
	"github.com/evo-lee/evo-sonic/handler/trans"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/dto"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/service/assembler"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type SheetHandler struct {
	SheetService   service.SheetService
	PostService    service.PostService
	SheetAssembler assembler.SheetAssembler
}

func NewSheetHandler(sheetService service.SheetService, postService service.PostService, sheetAssembler assembler.SheetAssembler) *SheetHandler {
	return &SheetHandler{
		SheetService:   sheetService,
		PostService:    postService,
		SheetAssembler: sheetAssembler,
	}
}

func (s *SheetHandler) GetSheetByID(ctx web.Context) (interface{}, error) {
	sheetID, err := util.ParamWebInt32(ctx, "sheetID")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	sheet, err := s.SheetService.GetByPostID(reqCtx, sheetID)
	if err != nil {
		return nil, err
	}
	return s.SheetAssembler.ConvertToDetailVO(reqCtx, sheet)
}

func (s *SheetHandler) ListSheet(ctx web.Context) (interface{}, error) {
	type SheetParam struct {
		param.Page
		Sort string `json:"sort"`
	}
	var sheetParam SheetParam
	err := ctx.BindQuery(&sheetParam)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("Parameter error")
	}
	reqCtx := ctx.RequestContext()
	sheets, totalCount, err := s.SheetService.Page(reqCtx, sheetParam.Page, &param.Sort{Fields: []string{"createTime,desc"}})
	if err != nil {
		return nil, err
	}
	sheetVOs, err := s.SheetAssembler.ConvertToListVO(reqCtx, sheets)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(sheetVOs, totalCount, sheetParam.Page), nil
}

func (s *SheetHandler) IndependentSheets(ctx web.Context) (interface{}, error) {
	return s.SheetService.ListIndependentSheets(ctx.RequestContext())
}

func (s *SheetHandler) CreateSheet(ctx web.Context) (interface{}, error) {
	var sheetParam param.Sheet
	err := ctx.BindJSON(&sheetParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest)
	}
	reqCtx := ctx.RequestContext()
	sheet, err := s.SheetService.Create(reqCtx, &sheetParam)
	if err != nil {
		return nil, err
	}
	sheetDetailVO, err := s.SheetAssembler.ConvertToDetailVO(reqCtx, sheet)
	if err != nil {
		return nil, err
	}
	return sheetDetailVO, nil
}

func (s *SheetHandler) UpdateSheet(ctx web.Context) (interface{}, error) {
	var sheetParam param.Sheet
	err := ctx.BindJSON(&sheetParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}

	sheetID, err := util.ParamWebInt32(ctx, "sheetID")
	if err != nil {
		return nil, err
	}
	postDetailVO, err := s.SheetService.Update(ctx.RequestContext(), sheetID, &sheetParam)
	if err != nil {
		return nil, err
	}
	return postDetailVO, nil
}

func (s *SheetHandler) UpdateSheetStatus(ctx web.Context) (interface{}, error) {
	sheetID, err := util.ParamWebInt32(ctx, "sheetID")
	if err != nil {
		return nil, err
	}
	statusStr, err := util.ParamWebString(ctx, "status")
	if err != nil {
		return nil, err
	}
	status, err := consts.PostStatusFromString(statusStr)
	if err != nil {
		return nil, err
	}
	if status < consts.PostStatusPublished || status > consts.PostStatusIntimate {
		return nil, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg("status error")
	}
	return s.SheetService.UpdateStatus(ctx.RequestContext(), sheetID, status)
}

func (s *SheetHandler) UpdateSheetDraft(ctx web.Context) (interface{}, error) {
	sheetID, err := util.ParamWebInt32(ctx, "sheetID")
	if err != nil {
		return nil, err
	}
	var postContentParam param.PostContent
	err = ctx.BindJSON(&postContentParam)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("content param error")
	}
	reqCtx := ctx.RequestContext()
	post, err := s.SheetService.UpdateDraftContent(reqCtx, sheetID, postContentParam.Content, postContentParam.OriginalContent)
	if err != nil {
		return nil, err
	}
	return s.SheetAssembler.ConvertToDetailDTO(reqCtx, post)
}

func (s *SheetHandler) DeleteSheet(ctx web.Context) (interface{}, error) {
	sheetID, err := util.ParamWebInt32(ctx, "sheetID")
	if err != nil {
		return nil, err
	}
	return nil, s.SheetService.Delete(ctx.RequestContext(), sheetID)
}

func (s *SheetHandler) PreviewSheet(ctx web.Context) {
	sheetID, err := util.ParamWebInt32(ctx, "sheetID")
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	previewPath, err := s.SheetService.Preview(ctx.RequestContext(), sheetID)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.String(http.StatusOK, previewPath)
}
