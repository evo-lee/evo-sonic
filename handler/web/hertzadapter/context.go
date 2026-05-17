package hertzadapter

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	hertzapp "github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	"github.com/cloudwego/hertz/pkg/protocol"

	"github.com/evo-lee/evo-sonic/handler/web"
)

const (
	requestContextKey   = "__sonic_request_context__"
	clientIPContextKey  = "client_ip"
	userAgentContextKey = "user_agent"
)

type Context struct {
	baseCtx      context.Context
	ctx          *hertzapp.RequestContext
	compatReq    *http.Request
	compatReqErr error
}

func NewContext(baseCtx context.Context, ctx *hertzapp.RequestContext) web.Context {
	if ctx != nil {
		if persisted, ok := ctx.Get(requestContextKey); ok {
			if persistedCtx, ok := persisted.(context.Context); ok && persistedCtx != nil {
				baseCtx = persistedCtx
			}
		}
	}
	if baseCtx == nil {
		baseCtx = context.Background()
	}
	if ctx != nil {
		if baseCtx.Value(clientIPContextKey) == nil {
			baseCtx = context.WithValue(baseCtx, clientIPContextKey, ctx.ClientIP())
		}
		if baseCtx.Value(userAgentContextKey) == nil {
			baseCtx = context.WithValue(baseCtx, userAgentContextKey, string(ctx.GetHeader("User-Agent")))
		}
	}
	return &Context{baseCtx: baseCtx, ctx: ctx}
}

func (c *Context) Deadline() (time.Time, bool) {
	return c.baseCtx.Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.baseCtx.Done()
}

func (c *Context) Err() error {
	return c.baseCtx.Err()
}

func (c *Context) Value(key any) any {
	return c.baseCtx.Value(key)
}

func (c *Context) RequestContext() context.Context {
	return c.baseCtx
}

func (c *Context) Request() *http.Request {
	if c.compatReq != nil || c.compatReqErr != nil {
		return c.compatReq
	}
	c.compatReq, c.compatReqErr = adaptor.GetCompatRequest(&c.ctx.Request)
	return c.compatReq
}

func (c *Context) Writer() io.Writer {
	return c.ctx
}

func (c *Context) Method() string {
	return string(c.ctx.Method())
}

func (c *Context) Path() string {
	return string(c.ctx.Path())
}

func (c *Context) RawQuery() string {
	return string(c.ctx.URI().QueryString())
}

func (c *Context) ClientIP() string {
	return c.ctx.ClientIP()
}

func (c *Context) Header(key string) string {
	return string(c.ctx.GetHeader(key))
}

func (c *Context) ResponseHeader(key string) string {
	return string(c.ctx.Response.Header.Peek(key))
}

func (c *Context) SetHeader(key, value string) {
	c.ctx.Header(key, value)
}

func (c *Context) StatusCode() int {
	return c.ctx.Response.StatusCode()
}

func (c *Context) Query(key string) (string, bool) {
	return c.ctx.GetQuery(key)
}

func (c *Context) Param(key string) string {
	return c.ctx.Param(key)
}

func (c *Context) Cookie(name string) (string, error) {
	value := c.ctx.Cookie(name)
	if len(value) == 0 {
		return "", http.ErrNoCookie
	}
	return string(value), nil
}

func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	c.ctx.SetCookie(name, value, maxAge, path, domain, protocol.CookieSameSiteDisabled, secure, httpOnly)
}

func (c *Context) MultipartForm() (*multipart.Form, error) {
	return c.ctx.MultipartForm()
}

func (c *Context) Set(key string, value any) {
	c.baseCtx = context.WithValue(c.baseCtx, key, value)
	c.ctx.Set(requestContextKey, c.baseCtx)
	c.ctx.Set(key, value)
}

func (c *Context) Get(key any) (any, bool) {
	if value := c.baseCtx.Value(key); value != nil {
		return value, true
	}
	strKey, ok := key.(string)
	if ok {
		return c.ctx.Get(strKey)
	}
	return nil, false
}

func (c *Context) Bind(value any) error {
	return c.ctx.Bind(value)
}

func (c *Context) BindJSON(value any) error {
	return c.ctx.BindJSON(value)
}

func (c *Context) BindQuery(value any) error {
	return c.ctx.BindQuery(value)
}

func (c *Context) BindWith(value any, binder any) error {
	requestBinder, ok := binder.(web.RequestBinder)
	if !ok {
		return fmt.Errorf("unsupported binder type %T", binder)
	}
	req := c.Request()
	if req == nil {
		if c.compatReqErr != nil {
			return c.compatReqErr
		}
		return fmt.Errorf("failed to build http request compatibility wrapper")
	}
	return requestBinder.Bind(req, value)
}

func (c *Context) JSON(status int, value any) {
	c.ctx.JSON(status, value)
}

func (c *Context) AbortWithStatusJSON(status int, value any) {
	c.ctx.AbortWithStatusJSON(status, value)
}

func (c *Context) String(status int, value string) {
	c.ctx.String(status, "%s", value)
}

func (c *Context) Status(status int) {
	c.ctx.Status(status)
}

func (c *Context) Redirect(status int, location string) {
	c.ctx.Redirect(status, []byte(location))
}

func (c *Context) File(path string) {
	c.ctx.File(path)
}

func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	return c.ctx.FormFile(name)
}

func (c *Context) Abort() {
	c.ctx.Abort()
}

func (c *Context) Next() {
	c.ctx.Next(c.baseCtx)
}

func (c *Context) Native() any {
	return c.ctx
}
