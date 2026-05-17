package vo

import "github.com/evo-lee/evo-sonic/model/dto"

type Menu struct {
	dto.Menu
	Children []*Menu `json:"children"`
}
