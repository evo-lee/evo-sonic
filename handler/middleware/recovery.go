package middleware

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/evo-lee/evo-sonic/handler/web"
)

type RecoveryMiddleware struct {
	logger *zap.Logger
}

func NewRecoveryMiddleware(logger *zap.Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{logger: logger}
}

func (r *RecoveryMiddleware) Handler() web.HandlerFunc {
	logger := r.logger.WithOptions(zap.AddCallerSkip(2))
	return func(ctx web.Context) {
		defer func() {
			if panicVal := recover(); panicVal != nil {
				var recoveredErr error
				switch e := panicVal.(type) {
				case error:
					recoveredErr = e
				default:
					recoveredErr = fmt.Errorf("%v", e)
				}
				var brokenPipe bool
				if ne, ok := recoveredErr.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						msg := strings.ToLower(se.Error())
						if strings.Contains(msg, "broken pipe") || strings.Contains(msg, "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				if brokenPipe {
					logger.Error(ctx.Path(), zap.Error(recoveredErr), zap.String("request_id", GetRequestID(ctx)))
				} else {
					logger.Error("[Recovery] panic recovered",
						zap.Error(recoveredErr),
						zap.String("request_id", GetRequestID(ctx)),
						zap.String("method", ctx.Method()),
						zap.String("path", ctx.Path()),
					)
				}

				if brokenPipe {
					ctx.Abort()
				} else {
					code := http.StatusInternalServerError
					AbortWithErrorJSON(ctx, code, ErrorCodeFromStatus(code), LocalizedHTTPStatusText(ctx, code))
				}
			}
		}()
		ctx.Next()
	}
}
