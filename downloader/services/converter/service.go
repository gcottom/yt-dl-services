package converter

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/gcottom/go-zaplog"
	"go.uber.org/zap"
)

func (s *Service) Convert(ctx context.Context, id string) error {
	var args = []string{"-i", fmt.Sprintf("%s/%s.temp", s.Config.TempDir, id), "-c:a", "libmp3lame", "-b:a", "256k", "-f", "mp3", fmt.Sprintf("./data/%s.mp3", id)}
	cmd := exec.Command(s.Config.FFMPEGPath, args...)

	zaplog.InfoC(ctx, "converting file", zap.String("id", id))
	err := cmd.Start() // Start a process on another goroutine
	if err != nil {
		zaplog.ErrorC(ctx, "conversion error", zap.Error(err))
		return err
	}

	err = cmd.Wait() // wait until ffmpeg finish
	if err != nil {
		zaplog.ErrorC(ctx, "conversion error", zap.Error(err))
		return err
	}
	zaplog.InfoC(ctx, "conversion complete", zap.String("id", id))
	if err := os.Remove(fmt.Sprintf("%s/%s.temp", s.Config.TempDir, id)); err != nil {
		zaplog.ErrorC(ctx, "failed to remove temp file", zap.Error(err))
		return err
	}
	return nil
}
