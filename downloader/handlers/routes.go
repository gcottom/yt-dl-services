package handlers

import (
	"github.com/gcottom/yt-dl-services/downloader/services/download"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, downloadService download.DownloadService) {
	h := &Handler{DownloadService: downloadService}

	router.Group("/api").
		GET("/download", h.StartDownload).
		GET("/status", func(ctx *gin.Context) {})
}
