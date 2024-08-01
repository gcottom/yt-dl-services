package download

import (
	"context"

	"github.com/gcottom/yt-dl-services/downloader/config"
	"github.com/gcottom/yt-dl-services/downloader/pkg/http_client"
	"github.com/gcottom/yt-dl-services/downloader/services/converter"
	"github.com/gcottom/yt-dl-services/downloader/services/meta"
	"github.com/gcottom/yt-dl-services/downloader/services/youtube_v2"
	"github.com/gcottom/yt-dl-services/downloader/track_sql"
)

type DownloadService interface {
	InitiateDownload(ctx context.Context, id string) error
}

func NewDownloadService(cfg *config.Config, httpClient *http_client.HTTPClient, trackSQL *track_sql.Client) *Service {
	return &Service{
		Config:                       cfg,
		HTTPClient:                   httpClient,
		Converter:                    &converter.Service{Config: cfg},
		MetaService:                  &meta.Service{Config: cfg, HTTPClient: httpClient},
		DownloadQueue:                make(chan string, 100),
		YoutubeService:               youtube_v2.NewYoutubeService(cfg, httpClient),
		TrackSQL:                     trackSQL,
		DLConcurrencyLimiter:         make(chan struct{}, cfg.Concurrency.Download),
		ConversionConcurrencyLimiter: make(chan struct{}, cfg.Concurrency.Conversion),
		GenreConcurrencyLimiter:      make(chan struct{}, cfg.Concurrency.Genre),
		PlaylistStatus:               make(map[string]bool),
	}
}

type Service struct {
	Config                       *config.Config
	HTTPClient                   *http_client.HTTPClient
	Converter                    converter.ConverterService
	MetaService                  meta.MetaService
	DownloadQueue                chan string
	YoutubeService               youtube_v2.YoutubeService
	TrackSQL                     *track_sql.Client
	DLConcurrencyLimiter         chan struct{}
	ConversionConcurrencyLimiter chan struct{}
	GenreConcurrencyLimiter      chan struct{}
	PlaylistStatus               map[string]bool
}

type GenreResponse struct {
	Genre string `json:"genre"`
}
