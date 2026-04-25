package aiimpl

import (
	"github.com/go-sonic/sonic/injection"
)

func init() {
	injection.Provide(
		NewConfigurableProvider,
		NewContentService,
		NewConfigurableEmbeddingProvider,
	)
}
