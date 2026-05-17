package middleware

import (
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/evo-lee/evo-sonic/handler/web"
)

type LoggerMiddleware struct {
	logger *zap.Logger
}

func NewLoggerMiddleware(logger *zap.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{logger: logger}
}

// LoggerConfig defines the config for Logger middleware.
type LoggerConfig struct {
	SkipPaths []string
}

func (g *LoggerMiddleware) HandlerWithConfig(conf LoggerConfig) web.HandlerFunc {
	logger := g.logger.WithOptions(zap.WithCaller(false))
	notLogged := conf.SkipPaths

	var skip map[string]struct{}
	if length := len(notLogged); length > 0 {
		skip = make(map[string]struct{}, length)
		for _, path := range notLogged {
			skip[path] = struct{}{}
		}
	}

	return func(ctx web.Context) {
		start := time.Now()
		path := ctx.Path()
		raw := ctx.RawQuery()
		ctx.Next()
		if _, ok := skip[path]; !ok {
			if raw != "" {
				path = path + "?" + raw
			}
			path = strings.ReplaceAll(path, "\n", "")
			path = strings.ReplaceAll(path, "\r", "")
			clientIP := strings.ReplaceAll(ctx.ClientIP(), "\n", "")
			clientIP = strings.ReplaceAll(clientIP, "\r", "")

			logger.Info("[HTTP]",
				zap.Time("beginTime", start),
				zap.Int("status", ctx.StatusCode()),
				zap.Duration("latency", time.Since(start)),
				zap.String("clientIP", clientIP),
				zap.String("method", ctx.Method()),
				zap.String("path", path),
				zap.String("request_id", GetRequestID(ctx)))
		}
	}
}
