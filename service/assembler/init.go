package assembler

import "github.com/evo-lee/evo-sonic/injection"

func init() {
	injection.Provide(
		NewBasePostAssembler,
		NewPostAssembler,
		NewSheetAssembler,
		NewBaseCommentAssembler,
		NewPostCommentAssembler,
		NewJournalCommentAssembler,
		NewSheetCommentAssembler,
	)
}
