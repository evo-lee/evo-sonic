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

type MenuHandler struct {
	MenuService service.MenuService
}

func NewMenuHandler(menuService service.MenuService) *MenuHandler {
	return &MenuHandler{
		MenuService: menuService,
	}
}

func (m *MenuHandler) ListMenus(ctx web.Context) (interface{}, error) {
	sort := param.Sort{}
	err := ctx.BindQuery(&sort)
	if err != nil {
		return nil, xerr.WithMsg(err, "sort parameter error").WithStatus(xerr.StatusBadRequest)
	}
	if len(sort.Fields) == 0 {
		sort.Fields = append(sort.Fields, "team,desc", "priority,asc")
	} else {
		sort.Fields = append(sort.Fields, "priority,asc")
	}
	reqCtx := ctx.RequestContext()
	menus, err := m.MenuService.List(reqCtx, &sort)
	if err != nil {
		return nil, err
	}
	return m.MenuService.ConvertToDTOs(reqCtx, menus), nil
}

func (m *MenuHandler) ListMenusAsTree(ctx web.Context) (interface{}, error) {
	sort := param.Sort{}
	err := ctx.BindQuery(&sort)
	if err != nil {
		return nil, xerr.WithMsg(err, "sort parameter error").WithStatus(xerr.StatusBadRequest)
	}
	if len(sort.Fields) == 0 {
		sort.Fields = append(sort.Fields, "team,desc", "priority,asc")
	} else {
		sort.Fields = append(sort.Fields, "priority,asc")
	}
	menus, err := m.MenuService.ListAsTree(ctx.RequestContext(), &sort)
	if err != nil {
		return nil, err
	}
	return menus, nil
}

func (m *MenuHandler) ListMenusAsTreeByTeam(ctx web.Context) (interface{}, error) {
	sort := param.Sort{}
	err := ctx.BindQuery(&sort)
	if err != nil {
		return nil, xerr.WithMsg(err, "sort parameter error").WithStatus(xerr.StatusBadRequest)
	}
	if len(sort.Fields) == 0 {
		sort.Fields = append(sort.Fields, "priority,asc")
	}
	team, _ := util.MustGetWebQueryString(ctx, "team")
	if team == "" {
		menus, err := m.MenuService.ListAsTree(ctx.RequestContext(), &sort)
		if err != nil {
			return nil, err
		}
		return menus, nil
	}
	menus, err := m.MenuService.ListAsTreeByTeam(ctx.RequestContext(), team, &sort)
	if err != nil {
		return nil, err
	}
	return menus, nil
}

func (m *MenuHandler) GetMenuByID(ctx web.Context) (interface{}, error) {
	id, err := util.ParamWebInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	reqCtx := ctx.RequestContext()
	menu, err := m.MenuService.GetByID(reqCtx, id)
	if err != nil {
		return nil, err
	}
	return m.MenuService.ConvertToDTO(reqCtx, menu), nil
}

func (m *MenuHandler) CreateMenu(ctx web.Context) (interface{}, error) {
	menuParam := &param.Menu{}
	err := ctx.BindJSON(menuParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	reqCtx := ctx.RequestContext()
	menu, err := m.MenuService.Create(reqCtx, menuParam)
	if err != nil {
		return nil, err
	}
	return m.MenuService.ConvertToDTO(reqCtx, menu), nil
}

func (m *MenuHandler) CreateMenuBatch(ctx web.Context) (interface{}, error) {
	menuParams := make([]*param.Menu, 0)
	err := ctx.BindJSON(&menuParams)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	reqCtx := ctx.RequestContext()
	menus, err := m.MenuService.CreateBatch(reqCtx, menuParams)
	if err != nil {
		return nil, err
	}
	return m.MenuService.ConvertToDTOs(reqCtx, menus), nil
}

func (m *MenuHandler) UpdateMenu(ctx web.Context) (interface{}, error) {
	id, err := util.ParamWebInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	menuParam := &param.Menu{}
	err = ctx.BindJSON(menuParam)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	reqCtx := ctx.RequestContext()
	menu, err := m.MenuService.Update(reqCtx, id, menuParam)
	if err != nil {
		return nil, err
	}
	return m.MenuService.ConvertToDTO(reqCtx, menu), nil
}

func (m *MenuHandler) UpdateMenuBatch(ctx web.Context) (interface{}, error) {
	menuParams := make([]*param.Menu, 0)
	err := ctx.BindJSON(&menuParams)
	if err != nil {
		e := validator.ValidationErrors{}
		if errors.As(err, &e) {
			return nil, xerr.WithStatus(e, xerr.StatusBadRequest).WithMsg(trans.Translate(e))
		}
		return nil, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg("parameter error")
	}
	reqCtx := ctx.RequestContext()
	menus, err := m.MenuService.UpdateBatch(reqCtx, menuParams)
	if err != nil {
		return nil, err
	}
	return m.MenuService.ConvertToDTOs(reqCtx, menus), nil
}

func (m *MenuHandler) DeleteMenu(ctx web.Context) (interface{}, error) {
	id, err := util.ParamWebInt32(ctx, "id")
	if err != nil {
		return nil, err
	}
	return nil, m.MenuService.Delete(ctx.RequestContext(), id)
}

func (m *MenuHandler) DeleteMenuBatch(ctx web.Context) (interface{}, error) {
	menuIDs := make([]int32, 0)
	err := ctx.BindJSON(&menuIDs)
	if err != nil {
		return nil, xerr.WithMsg(err, "menuIDs error").WithStatus(xerr.StatusBadRequest)
	}
	return nil, m.MenuService.DeleteBatch(ctx.RequestContext(), menuIDs)
}

func (m *MenuHandler) ListMenuTeams(ctx web.Context) (interface{}, error) {
	return m.MenuService.ListTeams(ctx.RequestContext())
}
