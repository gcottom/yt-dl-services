package handlers

import (
	"errors"
	"fmt"

	"github.com/gcottom/go-zaplog"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h *Handler) StartDownload(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		zaplog.WarnC(ctx, "start download request without ID present: ID is required")
		ResponseFailure(ctx, errors.New("start download request without ID present: ID is required"))
		return
	}
	zaplog.InfoC(ctx, "start download request received", zap.String("id", id))

	if err := h.DownloadService.InitiateDownload(ctx, id); err != nil {
		zaplog.ErrorC(ctx, "failed to start download", zap.String("id", id), zap.Error(err))
		ResponseInternalError(ctx, fmt.Errorf("failed to start download: %w", err))
		return
	}

	zaplog.InfoC(ctx, "start download request queued successfully", zap.String("id", id))
	ResponseSuccess(ctx, StartDownloadResponse{State: "ACK"})
}
