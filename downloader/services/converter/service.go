package converter

import (
	"bytes"
	"context"
	"os"
	"os/exec"

	"github.com/gcottom/go-zaplog"
	"go.uber.org/zap"
)

func (s *Service) Convert(ctx context.Context, b []byte) ([]byte, error) {
	var args = []string{"-i", "pipe:0", "-acodec", "aac", "-b:a", "192k", "-f", "ipod", "-"}
	cmd := exec.Command(s.Config.FFMPEGPath, args...)
	resultBuffer := bytes.NewBuffer(make([]byte, 0)) // pre allocate 5MiB buffer

	cmd.Stderr = os.Stderr    // bind log stream to stderr
	cmd.Stdout = resultBuffer // stdout result will be written here

	stdin, err := cmd.StdinPipe() // Open stdin pipe
	if err != nil {
		zaplog.ErrorC(ctx, "conversion error", zap.Error(err))
		return nil, err
	}

	err = cmd.Start() // Start a process on another goroutine
	if err != nil {
		zaplog.ErrorC(ctx, "conversion error", zap.Error(err))
		return nil, err
	}

	_, err = stdin.Write(b) // pump audio data to stdin pipe
	if err != nil {
		zaplog.ErrorC(ctx, "conversion error", zap.Error(err))
		return nil, err
	}
	err = stdin.Close() // close the stdin, or ffmpeg will wait forever
	if err != nil {
		zaplog.ErrorC(ctx, "conversion error", zap.Error(err))
		return nil, err
	}
	err = cmd.Wait() // wait until ffmpeg finish
	if err != nil {
		zaplog.ErrorC(ctx, "conversion error", zap.Error(err))
		return nil, err
	}
	return resultBuffer.Bytes(), nil
}
