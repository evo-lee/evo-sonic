package mcp

import "github.com/evo-lee/evo-sonic/injection"

func init() {
	injection.Provide(NewHandler)
}
