package gin_middleware

import (
	"context"
	"time"

	"github.com/gcottom/go-zaplog"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		zaplog.InfoC(ctx.Request.Context(), "request initiated")
		ctx.Next()
		latency := time.Since(start)
		statusCode := ctx.Writer.Status()
		zaplog.InfoC(ctx.Request.Context(), "request completed", zap.Int("status", statusCode), zap.Duration("latency", latency))
	}
}

func ContextMiddleware(baseCtx context.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Request = ctx.Request.WithContext(baseCtx)
		ctx.Next()
	}
}

func NewGinEngine(ctx context.Context) *gin.Engine {
	r := gin.New()
	r.Use(ContextMiddleware(ctx))
	r.Use(LoggingMiddleware())
	return r
}
