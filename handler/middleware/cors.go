package middleware

import (
	"strings"

	"github.com/evo-lee/evo-sonic/handler/web"
)

type CORSMiddleware struct {
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
}

func NewDevCORSMiddleware() *CORSMiddleware {
	return &CORSMiddleware{
		AllowMethods:     []string{"PUT", "PATCH", "GET", "DELETE", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Admin-Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}
}

func (c *CORSMiddleware) Handler() web.HandlerFunc {
	allowMethods := strings.Join(c.AllowMethods, ", ")
	allowHeaders := strings.Join(c.AllowHeaders, ", ")
	exposeHeaders := strings.Join(c.ExposeHeaders, ", ")
	return func(ctx web.Context) {
		origin := ctx.Header("Origin")
		if origin == "" {
			origin = "*"
		}
		ctx.SetHeader("Access-Control-Allow-Origin", origin)
		ctx.SetHeader("Vary", "Origin")
		ctx.SetHeader("Access-Control-Allow-Methods", allowMethods)
		ctx.SetHeader("Access-Control-Allow-Headers", allowHeaders)
		if exposeHeaders != "" {
			ctx.SetHeader("Access-Control-Expose-Headers", exposeHeaders)
		}
		if c.AllowCredentials {
			ctx.SetHeader("Access-Control-Allow-Credentials", "true")
		}
		if ctx.Method() == "OPTIONS" {
			ctx.Status(204)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
