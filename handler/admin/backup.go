package admin

import (
	"errors"
	"net/http"
	"path"
	"path/filepath"

	"github.com/go-playground/validator/v10"

	"github.com/evo-lee/evo-sonic/config"
	"github.com/evo-lee/evo-sonic/handler/trans"
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/log"
	"github.com/evo-lee/evo-sonic/model/dto"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type BackupHandler struct {
	BackupService service.BackupService
}

func NewBackupHandler(backupService service.BackupService) *BackupHandler {
	return &BackupHandler{
		BackupService: backupService,
	}
}

func (b *BackupHandler) GetWorkDirBackup(ctx web.Context) (interface{}, error) {
	filename, err := util.MustGetWebQueryString(ctx, "filename")
	if err != nil {
		return nil, err
	}
	return b.BackupService.GetBackup(ctx.RequestContext(), filepath.Join(config.BackupDir, filename), service.WholeSite)
}

func (b *BackupHandler) GetDataBackup(ctx web.Context) (interface{}, error) {
	filename, err := util.MustGetWebQueryString(ctx, "filename")
	if err != nil {
		return nil, err
	}
	return b.BackupService.GetBackup(ctx.RequestContext(), filepath.Join(config.DataExportDir, filename), service.JSONData)
}

func (b *BackupHandler) GetMarkDownBackup(ctx web.Context) (interface{}, error) {
	filename, err := util.MustGetWebQueryString(ctx, "filename")
	if err != nil {
		return nil, err
	}
	return b.BackupService.GetBackup(ctx.RequestContext(), filepath.Join(config.BackupMarkdownDir, filename), service.Markdown)
}

func (b *BackupHandler) BackupWholeSite(ctx web.Context) (interface{}, error) {
	toBackupItems := make([]string, 0)
	err := ctx.BindJSON(&toBackupItems)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest)
	}

	return b.BackupService.BackupWholeSite(ctx.RequestContext(), toBackupItems)
}

func (b *BackupHandler) ListBackups(ctx web.Context) (interface{}, error) {
	return b.BackupService.ListFiles(ctx.RequestContext(), config.BackupDir, service.WholeSite)
}

func (b *BackupHandler) ListToBackupItems(ctx web.Context) (interface{}, error) {
	return b.BackupService.ListToBackupItems(ctx.RequestContext())
}

func (b *BackupHandler) HandleWorkDir(ctx web.Context) {
	path := ctx.Path()
	if path == "/api/admin/backups/work-dir/fetch" {
		data, err := b.GetWorkDirBackup(ctx)
		respondWithJSONResult(ctx, data, err)
		return
	}
	if path == "/api/admin/backups/work-dir/options" || path == "/api/admin/backups/work-dir/options/" {
		data, err := b.ListToBackupItems(ctx)
		respondWithJSONResult(ctx, data, err)
		return
	}
	b.DownloadBackups(ctx)
}

func (b *BackupHandler) DownloadBackups(ctx web.Context) {
	filename := ctx.Param("path")
	if filename == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &dto.BaseDTO{
			Status:  http.StatusBadRequest,
			Message: "Filename parameter does not exist",
		})
		return
	}
	reqCtx := ctx.RequestContext()
	filePath, err := b.BackupService.GetBackupFilePath(reqCtx, config.BackupDir, filename)
	if err != nil {
		log.CtxErrorf(reqCtx, "err=%+v", err)
		status := xerr.GetHTTPStatus(err)
		ctx.JSON(status, &dto.BaseDTO{Status: status, Message: xerr.GetMessage(err)})
		return
	}
	ctx.File(filePath)
}

func (b *BackupHandler) DeleteBackups(ctx web.Context) (interface{}, error) {
	filename, err := util.MustGetWebQueryString(ctx, "filename")
	if err != nil {
		return nil, err
	}
	return nil, b.BackupService.DeleteFile(ctx.RequestContext(), config.BackupDir, filename)
}

func (b *BackupHandler) ImportMarkdown(ctx web.Context) (interface{}, error) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return nil, xerr.WithMsg(err, "上传文件错误").WithStatus(xerr.StatusBadRequest)
	}
	filenameExt := path.Ext(fileHeader.Filename)
	if filenameExt != ".md" && filenameExt != ".markdown" && filenameExt != ".mdown" {
		return nil, xerr.WithMsg(err, "Unsupported format").WithStatus(xerr.StatusBadRequest)
	}
	return nil, b.BackupService.ImportMarkdown(ctx.RequestContext(), fileHeader)
}

func (b *BackupHandler) ExportData(ctx web.Context) (interface{}, error) {
	return b.BackupService.ExportData(ctx.RequestContext())
}

func (b *BackupHandler) HandleData(ctx web.Context) {
	path := ctx.Path()
	if path == "/api/admin/backups/data/fetch" {
		data, err := b.GetDataBackup(ctx)
		respondWithJSONResult(ctx, data, err)
		return
	}
	if path == "/api/admin/backups/data" || path == "/api/admin/backups/data/" {
		data, err := b.ListExportData(ctx)
		respondWithJSONResult(ctx, data, err)
		return
	}
	b.DownloadData(ctx)
}

func (b *BackupHandler) ListExportData(ctx web.Context) (interface{}, error) {
	return b.BackupService.ListFiles(ctx.RequestContext(), config.DataExportDir, service.JSONData)
}

func (b *BackupHandler) DownloadData(ctx web.Context) {
	filename := ctx.Param("path")
	if filename == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &dto.BaseDTO{
			Status:  http.StatusBadRequest,
			Message: "Filename parameter does not exist",
		})
		return
	}
	reqCtx := ctx.RequestContext()
	filePath, err := b.BackupService.GetBackupFilePath(reqCtx, config.DataExportDir, filename)
	if err != nil {
		log.CtxErrorf(reqCtx, "err=%+v", err)
		status := xerr.GetHTTPStatus(err)
		ctx.JSON(status, &dto.BaseDTO{Status: status, Message: xerr.GetMessage(err)})
		return
	}
	ctx.File(filePath)
}

func (b *BackupHandler) DeleteDataFile(ctx web.Context) (interface{}, error) {
	filename, ok := ctx.Query("filename")
	if !ok || filename == "" {
		return nil, xerr.BadParam.New("no filename param").WithStatus(xerr.StatusBadRequest).WithMsg("no filename param")
	}
	return nil, b.BackupService.DeleteFile(ctx.RequestContext(), config.DataExportDir, filename)
}

func (b *BackupHandler) ExportMarkdown(ctx web.Context) (interface{}, error) {
	var exportMarkdownParam param.ExportMarkdown
	err := ctx.BindJSON(&exportMarkdownParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest)
	}
	return b.BackupService.ExportMarkdown(ctx.RequestContext(), exportMarkdownParam.NeedFrontMatter)
}

func (b *BackupHandler) ListMarkdowns(ctx web.Context) (interface{}, error) {
	return b.BackupService.ListFiles(ctx.RequestContext(), config.BackupMarkdownDir, service.Markdown)
}

func (b *BackupHandler) DeleteMarkdowns(ctx web.Context) (interface{}, error) {
	filename, err := util.MustGetWebQueryString(ctx, "filename")
	if err != nil {
		return nil, err
	}
	return nil, b.BackupService.DeleteFile(ctx.RequestContext(), config.BackupMarkdownDir, filename)
}

func (b *BackupHandler) DownloadMarkdown(ctx web.Context) {
	filename := ctx.Param("filename")
	if filename == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, &dto.BaseDTO{
			Status:  http.StatusBadRequest,
			Message: "Filename parameter does not exist",
		})
		return
	}
	reqCtx := ctx.RequestContext()
	filePath, err := b.BackupService.GetBackupFilePath(reqCtx, config.BackupMarkdownDir, filename)
	if err != nil {
		log.CtxErrorf(reqCtx, "err=%+v", err)
		status := xerr.GetHTTPStatus(err)
		ctx.JSON(status, &dto.BaseDTO{Status: status, Message: xerr.GetMessage(err)})
		return
	}
	ctx.File(filePath)
}

func respondWithJSONResult(ctx web.Context, data interface{}, err error) {
	if err != nil {
		log.CtxErrorf(ctx.RequestContext(), "err=%+v", err)
		status := xerr.GetHTTPStatus(err)
		ctx.JSON(status, &dto.BaseDTO{Status: status, Message: xerr.GetMessage(err)})
		return
	}

	ctx.JSON(http.StatusOK, &dto.BaseDTO{
		Status:  http.StatusOK,
		Data:    data,
		Message: "OK",
	})
}
