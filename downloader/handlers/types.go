package handlers

import "github.com/gcottom/yt-dl-services/downloader/services/download"

type Handler struct {
	DownloadService download.DownloadService
}
