package api

import "github.com/evo-lee/evo-sonic/injection"

func init() {
	injection.Provide(
		NewArchiveHandler,
		NewCategoryHandler,
		NewJournalHandler,
		NewLinkHandler,
		NewPostHandler,
		NewSheetHandler,
		NewOptionHandler,
		NewPhotoHandler,
		NewCommentHandler,
		NewSemanticSearchHandler,
	)
}
