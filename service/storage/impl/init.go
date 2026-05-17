package filestorageimpl

import "github.com/evo-lee/evo-sonic/injection"

func init() {
	injection.Provide(
		NewMinIO,
		NewLocalFileStorage,
		NewAliyun,
	)
}
