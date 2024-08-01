package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gcottom/go-zaplog"
	"github.com/gcottom/yt-dl-services/downloader/config"
	"github.com/gcottom/yt-dl-services/downloader/handlers"
	"github.com/gcottom/yt-dl-services/downloader/pkg/gin_middleware"
	"github.com/gcottom/yt-dl-services/downloader/pkg/http_client"
	"github.com/gcottom/yt-dl-services/downloader/services/download"
	"github.com/gcottom/yt-dl-services/downloader/track_sql"
)

func main() {
	config, err := config.LoadConfigFromFile("")
	if err != nil {
		panic(err)
	}
	if err := RunServer(config); err != nil {
		panic(err)
	}
}

func RunServer(cfg *config.Config) error {
	ctx := zaplog.CreateAndInject(context.Background())
	zaplog.InfoC(ctx, "starting downloader server")

	zaplog.InfoC(ctx, "creating http client")
	httpClient := http_client.NewHTTPClient()

	zaplog.InfoC(ctx, "creating track sql client")
	trackSQL, err := track_sql.NewClient(cfg)
	if err != nil {
		return err
	}

	zaplog.InfoC(ctx, "creating download service")
	downloadService := download.NewDownloadService(cfg, httpClient, trackSQL)

	zaplog.InfoC(ctx, "creating gin engine")
	ginws := gin_middleware.NewGinEngine(ctx)

	zaplog.InfoC(ctx, "setting up routes")
	handlers.SetupRoutes(ginws, downloadService)

	zaplog.InfoC(ctx, fmt.Sprintf("serving on port %d", cfg.Ports.Downloader))
	return http.ListenAndServe(fmt.Sprintf(":%d", cfg.Ports.Downloader), ginws)

}
