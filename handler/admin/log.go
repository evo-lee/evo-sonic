package admin

import (
	"github.com/evo-lee/evo-sonic/handler/web"
	"github.com/evo-lee/evo-sonic/model/dto"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/service"
	"github.com/evo-lee/evo-sonic/util"
	"github.com/evo-lee/evo-sonic/util/xerr"
)

type LogHandler struct {
	LogService service.LogService
}

func NewLogHandler(logService service.LogService) *LogHandler {
	return &LogHandler{
		LogService: logService,
	}
}

func (l *LogHandler) PageLatestLog(ctx web.Context) (interface{}, error) {
	top, err := util.MustGetWebQueryInt32(ctx, "top")
	if err != nil {
		top = 10
	}
	reqCtx := ctx.RequestContext()
	logs, _, err := l.LogService.PageLog(reqCtx, param.Page{PageSize: int(top)}, &param.Sort{Fields: []string{"createTime,desc"}})
	if err != nil {
		return nil, err
	}
	logDTOs := make([]*dto.Log, 0, len(logs))
	for _, log := range logs {
		logDTOs = append(logDTOs, l.LogService.ConvertToDTO(log))
	}
	return logDTOs, nil
}

func (l *LogHandler) PageLog(ctx web.Context) (interface{}, error) {
	type LogParam struct {
		param.Page
		*param.Sort
	}
	var logParam LogParam
	err := ctx.BindQuery(&logParam)
	if err != nil {
		return nil, xerr.WithMsg(err, "parameter error").WithStatus(xerr.StatusBadRequest)
	}
	if logParam.Sort == nil {
		logParam.Sort = &param.Sort{
			Fields: []string{"createTime,desc"},
		}
	}
	reqCtx := ctx.RequestContext()
	logs, totalCount, err := l.LogService.PageLog(reqCtx, logParam.Page, logParam.Sort)
	if err != nil {
		return nil, err
	}
	logDTOs := make([]*dto.Log, 0, len(logs))
	for _, log := range logs {
		logDTOs = append(logDTOs, l.LogService.ConvertToDTO(log))
	}
	return dto.NewPage(logDTOs, totalCount, logParam.Page), nil
}

func (l *LogHandler) ClearLog(ctx web.Context) (interface{}, error) {
	return nil, l.LogService.Clear(ctx.RequestContext())
}
