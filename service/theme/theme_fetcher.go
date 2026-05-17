package theme

import (
	"context"

	"github.com/evo-lee/evo-sonic/model/dto"
)

type ThemeFetcher interface {
	FetchTheme(ctx context.Context, file interface{}) (*dto.ThemeProperty, error)
}
