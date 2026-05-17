package aiimpl

import (
	"github.com/evo-lee/evo-sonic/injection"
)

func init() {
	injection.Provide(
		NewConfigurableProvider,
		NewContentService,
		NewConfigurableEmbeddingProvider,
	)
}
