package contextconvert

import (
	"context"

	"github.com/gin-gonic/gin"
)

func ConvertContext(ginctx *gin.Context) context.Context {
	ctx := ginctx.Request.Context()
	logger, exists := ginctx.Get("logger")
	if exists {
		ctx = context.WithValue(ctx, "logger", logger)
	}
	return ctx
}
