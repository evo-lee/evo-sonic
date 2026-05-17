package admin

import "github.com/evo-lee/evo-sonic/injection"

func init() {
	injection.Provide(
		NewAIHandler,
		NewAdminHandler,
		NewAttachmentHandler,
		NewCategoryHandler,
		NewBackupHandler,
		NewInstallHandler,
		NewJournalHandler,
		NewJournalCommentHandler,
		NewLinkHandler,
		NewLogHandler,
		NewMenuHandler,
		NewOptionHandler,
		NewPhotoHandler,
		NewPostHandler,
		NewPostCommentHandler,
		NewSheetHandler,
		NewSheetCommentHandler,
		NewStatisticHandler,
		NewTagHandler,
		NewThemeHandler,
		NewUserHandler,
		NewEmailHandler,
	)
}
