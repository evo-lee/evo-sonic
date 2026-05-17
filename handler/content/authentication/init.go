package authentication

import "github.com/evo-lee/evo-sonic/injection"

func init() {
	injection.Provide(
		NewCategoryAuthentication,
		NewPostAuthentication,
	)
}
