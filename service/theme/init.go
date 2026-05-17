package theme

import (
	"go.uber.org/fx"

	"github.com/evo-lee/evo-sonic/injection"
)

func init() {
	injection.Provide(
		NewFileScanner,
		NewPropertyScanner,
		fx.Annotated{Target: NewMultipartZipThemeFetcher, Name: "multipartZipThemeFetcher"},
		fx.Annotated{Target: NewGitThemeFetcher, Name: "gitRepoThemeFetcher"},
		fx.Annotated{Target: NewURLZipThemeFetcher, Name: "urlZipThemeFetcher"},
	)
}
