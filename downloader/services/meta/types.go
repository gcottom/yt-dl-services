package meta

import (
	"context"

	"github.com/gcottom/yt-dl-services/downloader/config"
	"github.com/gcottom/yt-dl-services/downloader/pkg/http_client"
	"github.com/gcottom/yt-dl-services/downloader/track_sql"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type MetaService interface {
	CoverArtistCheck(ctx context.Context, str string) string
	EqualIgnoringWhitespace(s1, s2 string) bool
	GetBestMetaMatch(ctx context.Context, trackMeta TrackMeta, spotifyMetas []TrackMeta) TrackMeta
	GetSpotifyToken(ctx context.Context) (*oauth2.Token, error)
	GetSpotifyMeta(ctx context.Context, trackMeta TrackMeta) ([]TrackMeta, error)
	GetYTMetaFromID(ctx context.Context, trackData track_sql.Track) (TrackMeta, error)
	SaveMeta(ctx context.Context, data []byte, trackData track_sql.Track) ([]byte, TrackMeta, error)
	SanitizeAuthor(author string) string
	SanitizeParenthesis(str string) string
	SanitizeString(str string) string
}

type Service struct {
	Config        *config.Config
	HTTPClient    *http_client.HTTPClient
	SpotifyConfig *clientcredentials.Config
}

type TrackMeta struct {
	Title       string
	Artist      string
	Album       string
	Genre       string
	CoverArtURL string
}

type YTMMetaResponse struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Image  string `json:"image"`
	Type   string `json:"type"`
}
