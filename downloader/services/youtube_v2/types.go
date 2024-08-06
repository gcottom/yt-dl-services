package youtube_v2

import (
	"context"

	"github.com/gcottom/semaphore"
	"github.com/gcottom/yt-dl-services/downloader/config"
	"github.com/gcottom/yt-dl-services/downloader/pkg/http_client"
	"github.com/kkdai/youtube/v2"
)

type YoutubeService interface {
	Download(ctx context.Context, id string, useEmbedded bool) ([]byte, error)
	GetPlaylistEntries(ctx context.Context, playlistID string) ([]string, error)
	GetVideoInfo(ctx context.Context, videoID string, useEmbedded bool) (string, string, error)
}

type Service struct {
	Config           *config.Config
	HTTPClient       *http_client.HTTPClient
	YTClient         *youtube.Client
	YTEmbeddedClient *youtube.Client
	DLCount          int
	DLRateSemaphore  *semaphore.Semaphore
}

func NewYoutubeService(config *config.Config, httpClient *http_client.HTTPClient) *Service {
	youtube.DefaultClient = youtube.EmbeddedClient
	embeddedClient := &youtube.Client{HTTPClient: httpClient.Client}
	embeddedClient.GetVideo("0P19rsu3jXY")
	youtube.DefaultClient = youtube.AndroidClient
	androidClient := &youtube.Client{HTTPClient: httpClient.Client}
	androidClient.GetVideo("0P19rsu3jXY")
	return &Service{
		Config:           config,
		HTTPClient:       httpClient,
		YTClient:         androidClient,
		YTEmbeddedClient: embeddedClient,
		DLCount:          0,
		DLRateSemaphore:  semaphore.NewSemaphore(config.Concurrency.Download),
	}
}

func (s *Service) ReInitializeClients() {
	s.HTTPClient.Client.CloseIdleConnections()
	httpClient := http_client.NewHTTPClient()
	s.HTTPClient = httpClient
	youtube.DefaultClient = youtube.EmbeddedClient
	embeddedClient := &youtube.Client{HTTPClient: httpClient.Client}
	embeddedClient.GetVideo("0P19rsu3jXY")
	youtube.DefaultClient = youtube.AndroidClient
	androidClient := &youtube.Client{HTTPClient: httpClient.Client}
	androidClient.GetVideo("0P19rsu3jXY")
	s.YTClient = androidClient
	s.YTEmbeddedClient = embeddedClient
}
