package download

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gcottom/go-zaplog"
	"github.com/gcottom/yt-dl-services/downloader/services/meta"
	"github.com/gcottom/yt-dl-services/downloader/track_sql"
	"go.uber.org/zap"
)

func (s *Service) InitiateDownload(ctx context.Context, id string) error {
	s.DownloadQueue <- id
	return nil
}

func (s *Service) IsTrackID(id string) bool {
	return len(id) == 11
}

func (s *Service) GetDownloadStatus(ctx context.Context, id string) (string, error) {
	return "", nil
}

func (s *Service) QueueProcessor(ctx context.Context) {
	for {
		select {
		case id := <-s.DownloadQueue:
			s.processDownload(ctx, id)
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func (s *Service) processDownload(ctx context.Context, id string) {
	zaplog.InfoC(ctx, "processing download", zap.String("id", id))
	if s.IsTrackID(id) {
		zaplog.InfoC(ctx, "given ID is a track ID", zap.String("id", id))
		wg := new(sync.WaitGroup)
		wg.Add(1)
		s.DLConcurrencyLimiter <- struct{}{}
		go s.processTrack(ctx, id, wg)
	} else {
		zaplog.InfoC(ctx, "given ID is a playlist ID", zap.String("id", id))
		go s.processPlaylist(ctx, id)
	}
}

func (s *Service) processPlaylist(ctx context.Context, id string) {
	s.PlaylistStatus[id] = false
	playlistEntries, err := s.YoutubeService.GetPlaylistEntries(ctx, id)
	if err != nil {
		s.PlaylistStatus[id] = true
		zaplog.ErrorC(ctx, "failed to get playlist entries", zap.String("id", id), zap.Error(err))
		return
	}
	wg := new(sync.WaitGroup)
	for _, entry := range playlistEntries {
		wg.Add(1)
		s.DLConcurrencyLimiter <- struct{}{}
		go s.processTrack(ctx, entry, wg)
	}
	wg.Wait()
	s.PlaylistStatus[id] = true
}

func (s *Service) processTrack(ctx context.Context, id string, wg *sync.WaitGroup) error {
	zaplog.InfoC(ctx, "processing track", zap.String("id", id))
	track, trackData, err := s.retrieveTrack(ctx, id)
	if err != nil {
		zaplog.ErrorC(ctx, "failed to retrieve track", zap.String("id", id), zap.Error(err))
		<-s.DLConcurrencyLimiter
		wg.Done()
		return err
	}
	<-s.DLConcurrencyLimiter
	s.ConversionConcurrencyLimiter <- struct{}{}
	track, convertedData, err := s.convertTrack(ctx, trackData, track)
	if err != nil {
		zaplog.ErrorC(ctx, "failed to convert track", zap.String("id", id), zap.Error(err))
		<-s.ConversionConcurrencyLimiter
		wg.Done()
		return err
	}
	<-s.ConversionConcurrencyLimiter
	s.GenreConcurrencyLimiter <- struct{}{}
	track, err = s.getGenre(ctx, convertedData, track)
	if err != nil {
		zaplog.ErrorC(ctx, "failed to get genre", zap.String("id", id), zap.Error(err))
		<-s.GenreConcurrencyLimiter
		wg.Done()
		return err
	}
	<-s.GenreConcurrencyLimiter
	outputData, meta, err := s.MetaService.SaveMeta(ctx, convertedData, track)
	if err != nil {
		zaplog.ErrorC(ctx, "failed to save meta", zap.String("id", id), zap.Error(err))
		wg.Done()
		return err
	}
	if err := s.saveFile(ctx, outputData, meta); err != nil {
		zaplog.ErrorC(ctx, "failed to save file", zap.String("id", id), zap.Error(err))
		track.Error = 1
		track.ErrorMessage = err.Error()
		if err := s.TrackSQL.UpdateTrack(ctx, track); err != nil {
			zaplog.ErrorC(ctx, "failed to insert track into db", zap.String("id", id), zap.Error(err))
		}
		wg.Done()
		return err
	}
	track.Done = 1
	if err := s.TrackSQL.UpdateTrack(ctx, track); err != nil {
		zaplog.ErrorC(ctx, "failed to insert track into db", zap.String("id", id), zap.Error(err))
	}
	wg.Done()
	return nil
}

func (s *Service) retrieveTrack(ctx context.Context, id string) (track_sql.Track, []byte, error) {
	var track track_sql.Track
	var err error
	var trackData []byte
	track.ID = id

	track.Title, track.Author, err = s.YoutubeService.GetVideoInfo(ctx, id, false)
	if err != nil {
		zaplog.ErrorC(ctx, "failed to get track info, attempting with embedded player", zap.String("id", id), zap.Error(err))
		track.Title, track.Author, err = s.YoutubeService.GetVideoInfo(ctx, id, true)
		if err != nil {
			track.Error = 1
			track.ErrorMessage = err.Error()
			zaplog.ErrorC(ctx, "failed to get track info with embedded player", zap.String("id", id), zap.Error(err))
			if err := s.TrackSQL.InsertTrack(ctx, track); err != nil {
				zaplog.ErrorC(ctx, "failed to insert track into db", zap.String("id", id), zap.Error(err))
			}
			return track, nil, err
		}
	}
	if err := s.TrackSQL.InsertTrack(ctx, track); err != nil {
		zaplog.ErrorC(ctx, "failed to insert track into db", zap.String("id", id), zap.Error(err))
		return track, nil, err
	}

	trackData, err = s.YoutubeService.Download(ctx, id, false)
	if err != nil {
		track.Error = 1
		track.ErrorMessage = err.Error()
		zaplog.ErrorC(ctx, "failed to download track", zap.String("id", id), zap.Error(err))
		if err := s.TrackSQL.UpdateTrack(ctx, track); err != nil {
			zaplog.ErrorC(ctx, "failed to insert track into db", zap.String("id", id), zap.Error(err))
		}
		return track, nil, err
	}
	return track, trackData, nil
}

func (s *Service) convertTrack(ctx context.Context, trackData []byte, track track_sql.Track) (track_sql.Track, []byte, error) {
	convertedData, err := s.Converter.Convert(ctx, trackData)
	if err != nil {
		track.Error = 1
		track.ErrorMessage = err.Error()
		zaplog.ErrorC(ctx, "failed to convert track", zap.String("id", track.ID), zap.Error(err))
		if err := s.TrackSQL.UpdateTrack(ctx, track); err != nil {
			zaplog.ErrorC(ctx, "failed to insert track into db", zap.String("id", track.ID), zap.Error(err))
		}
		return track, nil, err
	}
	return track, convertedData, nil
}

func (s *Service) getGenre(ctx context.Context, convertedData []byte, track track_sql.Track) (track_sql.Track, error) {
	req, err := s.HTTPClient.CreateOctetStreamRequest(http.MethodPost, fmt.Sprintf("http://localhost:%d%s", s.Config.Ports.Genre, s.Config.Endpoints.Genre), convertedData)
	if err != nil {
		track.Error = 1
		track.ErrorMessage = err.Error()
		zaplog.ErrorC(ctx, "failed to create genre request", zap.String("id", track.ID), zap.Error(err))
		if err := s.TrackSQL.UpdateTrack(ctx, track); err != nil {
			zaplog.ErrorC(ctx, "failed to insert track into db", zap.String("id", track.ID), zap.Error(err))
		}
		return track, err
	}

	resp, status, err := s.HTTPClient.DoRequest(req)
	if err != nil || status != http.StatusOK {
		track.Error = 1
		track.ErrorMessage = err.Error()
		zaplog.ErrorC(ctx, "failed to send genre request", zap.String("id", track.ID), zap.Error(err))
		if err := s.TrackSQL.UpdateTrack(ctx, track); err != nil {
			zaplog.ErrorC(ctx, "failed to insert track into db", zap.String("id", track.ID), zap.Error(err))
		}
		return track, err
	}
	var genreResponse GenreResponse
	if err := json.Unmarshal(resp, &genreResponse); err != nil {
		track.Error = 1
		track.ErrorMessage = err.Error()
		zaplog.ErrorC(ctx, "failed to unmarshal genre response", zap.String("id", track.ID), zap.Error(err))
		if err := s.TrackSQL.UpdateTrack(ctx, track); err != nil {
			zaplog.ErrorC(ctx, "failed to insert track into db", zap.String("id", track.ID), zap.Error(err))
		}
		return track, err
	}
	track.Genre = genreResponse.Genre
	if err := s.TrackSQL.UpdateTrack(ctx, track); err != nil {
		zaplog.ErrorC(ctx, "failed to insert track into db", zap.String("id", track.ID), zap.Error(err))
		return track, err
	}
	return track, nil
}

func (s *Service) saveFile(ctx context.Context, data []byte, meta meta.TrackMeta) error {
	fileName := s.sanitizeFilename(fmt.Sprintf("%s - %s", meta.Artist, meta.Title))
	outputFile, err := os.Create(fmt.Sprintf("%s/%s.m4a", s.Config.DownloadDir, fileName))
	if err != nil {
		zaplog.ErrorC(ctx, "failed to create file", zap.String("filename", fileName), zap.Error(err))
		return err
	}
	defer outputFile.Close()
	if _, err := outputFile.Write(data); err != nil {
		zaplog.ErrorC(ctx, "failed to write to file", zap.String("filename", fileName), zap.Error(err))
		return err
	}
	return nil
}

func (s *Service) sanitizeFilename(str string) string {
	regex := regexp.MustCompile(`[\\/:*?"<>|\x00-\x1F]`)
	safeStr := regex.ReplaceAllString(str, "_")
	return strings.Trim(safeStr, " .")

}
