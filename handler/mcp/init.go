package mcp

import "github.com/go-sonic/sonic/injection"

func init() {
	injection.Provide(NewHandler)
}
