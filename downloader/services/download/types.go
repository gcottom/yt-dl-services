package download

import (
	"context"

	"github.com/gcottom/semaphore"
	"github.com/gcottom/yt-dl-services/downloader/config"
	"github.com/gcottom/yt-dl-services/downloader/pkg/http_client"
	"github.com/gcottom/yt-dl-services/downloader/services/converter"
	"github.com/gcottom/yt-dl-services/downloader/services/meta"
	"github.com/gcottom/yt-dl-services/downloader/services/redriver"
	"github.com/gcottom/yt-dl-services/downloader/services/youtube_v2"
	"github.com/gcottom/yt-dl-services/downloader/track_sql"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

type DownloadService interface {
	InitiateDownload(ctx context.Context, id string) error
}

func NewDownloadService(cfg *config.Config, httpClient *http_client.HTTPClient, trackSQL *track_sql.Client) *Service {
	return &Service{
		Config:                       cfg,
		HTTPClient:                   httpClient,
		Converter:                    &converter.Service{Config: cfg},
		MetaService:                  &meta.Service{Config: cfg, HTTPClient: httpClient, SpotifyConfig: &clientcredentials.Config{ClientID: cfg.Spotify.ClientID, ClientSecret: cfg.Spotify.ClientSecret, TokenURL: spotifyauth.TokenURL}},
		DownloadQueue:                make(chan string, 100),
		YoutubeService:               youtube_v2.NewYoutubeService(cfg, httpClient),
		TrackSQL:                     trackSQL,
		DLConcurrencyLimiter:         semaphore.NewSemaphore(cfg.Concurrency.Download),
		ConversionConcurrencyLimiter: semaphore.NewSemaphore(cfg.Concurrency.Conversion),
		GenreConcurrencyLimiter:      semaphore.NewSemaphore(cfg.Concurrency.Genre),
		PlaylistStatus:               make(map[string]bool),
		ReDriver:                     redriver.NewService(),
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
	DLConcurrencyLimiter         *semaphore.Semaphore
	ConversionConcurrencyLimiter *semaphore.Semaphore
	GenreConcurrencyLimiter      *semaphore.Semaphore
	PlaylistStatus               map[string]bool
	ReDriver                     redriver.ReDriverService
}

type GenreResponse struct {
	Genre string `json:"genre"`
}
