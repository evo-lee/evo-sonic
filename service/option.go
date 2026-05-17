package service

import (
	"context"

	"github.com/evo-lee/evo-sonic/consts"
	"github.com/evo-lee/evo-sonic/model/dto"
	"github.com/evo-lee/evo-sonic/model/param"
	"github.com/evo-lee/evo-sonic/model/property"
)

type OptionService interface {
	GetOrByDefaultWithErr(ctx context.Context, p property.Property, defaultValue interface{}) (interface{}, error)
	GetOrByDefault(ctx context.Context, p property.Property) interface{}
	GetBlogBaseURL(ctx context.Context) (string, error)
	IsEnabledAbsolutePath(ctx context.Context) (bool, error)
	GetPathSuffix(ctx context.Context) (string, error)
	GetArchivePrefix(ctx context.Context) (string, error)
	Save(ctx context.Context, optionMap map[string]string) error
	ListAllOption(ctx context.Context) ([]*dto.Option, error)
	GetLinksPrefix(ctx context.Context) (string, error)
	GetPhotoPrefix(ctx context.Context) (string, error)
	GetJournalPrefix(ctx context.Context) (string, error)
	GetActivatedThemeID(ctx context.Context) (string, error)
	GetPostPermalinkType(ctx context.Context) (consts.PostPermalinkType, error)
	GetSheetPermalinkType(ctx context.Context) (consts.SheetPermaLinkType, error)
	GetIndexPageSize(ctx context.Context) int
	GetPostSort(ctx context.Context) param.Sort
	GetPostSummaryLength(ctx context.Context) int
	GetCategoryPrefix(ctx context.Context) (string, error)
	GetTagPrefix(ctx context.Context) (string, error)
	GetLinkPrefix(ctx context.Context) (string, error)
	GetSheetPrefix(ctx context.Context) (string, error)
	GetAttachmentType(ctx context.Context) consts.AttachmentType
	GetAdminURLPath(ctx context.Context) (string, error)
}

type ClientOptionService interface {
	OptionService
}
