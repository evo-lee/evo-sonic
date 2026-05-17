package vo

import "github.com/evo-lee/evo-sonic/model/dto"

type CategoryVO struct {
	dto.CategoryDTO
	Children []*CategoryVO `json:"children"`
}
