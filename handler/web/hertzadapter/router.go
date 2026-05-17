package hertzadapter

import (
	"context"

	hertzapp "github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"

	"github.com/evo-lee/evo-sonic/handler/web"
)

type Router struct {
	router hertzRouter
}

type hertzRouter interface {
	Use(...hertzapp.HandlerFunc) route.IRoutes
	Group(string, ...hertzapp.HandlerFunc) *route.RouterGroup
	GET(string, ...hertzapp.HandlerFunc) route.IRoutes
	POST(string, ...hertzapp.HandlerFunc) route.IRoutes
	PUT(string, ...hertzapp.HandlerFunc) route.IRoutes
	DELETE(string, ...hertzapp.HandlerFunc) route.IRoutes
	StaticFS(string, *hertzapp.FS) route.IRoutes
}

func NewRouter(router hertzRouter) web.Router {
	return &Router{router: router}
}

func wrapHandlers(handlers ...web.HandlerFunc) []hertzapp.HandlerFunc {
	if len(handlers) == 0 {
		return nil
	}
	wrapped := make([]hertzapp.HandlerFunc, 0, len(handlers))
	for _, handler := range handlers {
		current := handler
		wrapped = append(wrapped, func(ctx context.Context, reqCtx *hertzapp.RequestContext) {
			current(NewContext(ctx, reqCtx))
		})
	}
	return wrapped
}

func (r *Router) Use(handlers ...web.HandlerFunc) {
	r.router.Use(wrapHandlers(handlers...)...)
}

func (r *Router) Group(relativePath string, handlers ...web.HandlerFunc) web.Router {
	return &Router{router: r.router.Group(relativePath, wrapHandlers(handlers...)...)}
}

func (r *Router) GET(relativePath string, handlers ...web.HandlerFunc) {
	r.router.GET(relativePath, wrapHandlers(handlers...)...)
}

func (r *Router) POST(relativePath string, handlers ...web.HandlerFunc) {
	r.router.POST(relativePath, wrapHandlers(handlers...)...)
}

func (r *Router) PUT(relativePath string, handlers ...web.HandlerFunc) {
	r.router.PUT(relativePath, wrapHandlers(handlers...)...)
}

func (r *Router) DELETE(relativePath string, handlers ...web.HandlerFunc) {
	r.router.DELETE(relativePath, wrapHandlers(handlers...)...)
}

func (r *Router) StaticFS(relativePath, root string) {
	r.router.StaticFS(relativePath, &hertzapp.FS{
		Root:        root,
		IndexNames:  []string{"index.html"},
		PathRewrite: hertzapp.NewPathSlashesStripper(1),
	})
}

func (r *Router) Native() any {
	return r.router
}
