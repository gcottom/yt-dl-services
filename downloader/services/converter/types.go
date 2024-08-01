package converter

import (
	"context"

	"github.com/gcottom/yt-dl-services/downloader/config"
)

type ConverterService interface {
	Convert(ctx context.Context, b []byte) ([]byte, error)
}

type Service struct {
	Config *config.Config
}
