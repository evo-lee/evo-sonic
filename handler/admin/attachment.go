package admin

import (
	"github.com/evo-lee/evo-sonic/handler/binding"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/dto"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type AttachmentHandler struct {
	AttachmentService service.AttachmentService
}

func NewAttachmentHandler(attachmentService service.AttachmentService) *AttachmentHandler {
	return &AttachmentHandler{
		AttachmentService: attachmentService,
	}
}

func (a *AttachmentHandler) QueryAttachment(ctx web.Context) (interface{}, error) {
	queryParam := &param.AttachmentQuery{}
	err := ctx.BindWith(queryParam, binding.CustomFormBinding)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("param error ")
	}
	reqCtx := ctx.RequestContext()
	attachments, totalCount, err := a.AttachmentService.Page(reqCtx, queryParam)
	if err != nil {
		return nil, err
	}
	attachmentDTOs, err := a.AttachmentService.ConvertToDTOs(reqCtx, attachments)
	if err != nil {
		return nil, err
	}
	return dto.NewPage(attachmentDTOs, totalCount, queryParam.Page), nil
}

func (a *AttachmentHandler) GetAttachmentByID(ctx web.Context) (interface{}, error) {
	id, err := util.ParamWebInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	if id < 0 {
		return nil, xerr.BadParam.New("id < 0").WithStatus(xerr.StatusBadRequest).WithMsg("param error")
	}
	return a.AttachmentService.GetAttachment(ctx.RequestContext(), id)
}

func (a *AttachmentHandler) UploadAttachment(ctx web.Context) (interface{}, error) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return nil, xerr.WithMsg(err, "上传文件错误").WithStatus(xerr.StatusBadRequest)
	}
	return a.AttachmentService.Upload(ctx.RequestContext(), fileHeader)
}

func (a *AttachmentHandler) UploadAttachments(ctx web.Context) (interface{}, error) {
	form, _ := ctx.MultipartForm()
	if len(form.File) == 0 {
		return nil, xerr.BadParam.New("empty files").WithStatus(xerr.StatusBadRequest).WithMsg("empty files")
	}
	files := form.File["files"]
	attachmentDTOs := make([]*dto.AttachmentDTO, 0)
	for _, file := range files {
		attachment, err := a.AttachmentService.Upload(ctx.RequestContext(), file)
		if err != nil {
			return nil, err
		}
		attachmentDTOs = append(attachmentDTOs, attachment)
	}
	return attachmentDTOs, nil
}

func (a *AttachmentHandler) UpdateAttachment(ctx web.Context) (interface{}, error) {
	id, err := util.ParamWebInt32(ctx, "id")
	if err != nil {
		return nil, err
	}

	updateParam := &param.AttachmentUpdate{}
	err = ctx.Bind(updateParam)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("param error ")
	}
	return a.AttachmentService.Update(ctx.RequestContext(), id, updateParam)
}

func (a *AttachmentHandler) DeleteAttachment(ctx web.Context) (interface{}, error) {
	id, err := util.ParamWebInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	return a.AttachmentService.Delete(ctx.RequestContext(), id)
}

func (a *AttachmentHandler) DeleteAttachmentInBatch(ctx web.Context) (interface{}, error) {
	ids := make([]int32, 0)
	err := ctx.BindJSON(&ids)
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	return a.AttachmentService.DeleteBatch(ctx.RequestContext(), ids)
}

func (a *AttachmentHandler) GetAllMediaType(ctx web.Context) (interface{}, error) {
	return a.AttachmentService.GetAllMediaTypes(ctx.RequestContext())
}

func (a *AttachmentHandler) GetAllTypes(ctx web.Context) (interface{}, error) {
	attachmentTypes, err := a.AttachmentService.GetAllTypes(ctx.RequestContext())
	if err != nil {
		return nil, err
	}
	return attachmentTypes, nil
}
