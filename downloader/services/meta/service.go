package meta

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image/png"
	"net/http"
	"regexp"
	"strings"

	"github.com/gcottom/go-zaplog"
	"github.com/gcottom/mp4meta"
	"github.com/gcottom/yt-dl-services/downloader/track_sql"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

func (s *Service) SaveMeta(ctx context.Context, data []byte, trackData track_sql.Track) ([]byte, TrackMeta, error) {
	tag, err := mp4meta.ReadMP4(bytes.NewReader(data))
	if err != nil {
		zaplog.ErrorC(ctx, "failed to read mp4", zap.Error(err))
		return nil, TrackMeta{}, err
	}
	trackMeta, err := s.GetYTMetaFromID(ctx, trackData)
	if err != nil {
		zaplog.ErrorC(ctx, "failed to get yt meta", zap.Error(err))
		return nil, TrackMeta{}, err
	}
	spotifyMetas, err := s.GetSpotifyMeta(ctx, trackMeta)
	if err != nil {
		zaplog.ErrorC(ctx, "failed to get spotify meta", zap.Error(err))
		return nil, TrackMeta{}, err
	}
	bestMeta := s.GetBestMetaMatch(ctx, trackMeta, spotifyMetas)

	tag.SetTitle(bestMeta.Title)
	tag.SetArtist(bestMeta.Artist)
	tag.SetAlbum(bestMeta.Album)
	tag.SetGenre(bestMeta.Genre)
	if bestMeta.CoverArtURL != "" {
		response, err := http.Get(bestMeta.CoverArtURL)
		if err != nil {
			zaplog.ErrorC(ctx, "failed to get cover art", zap.Error(err))
			return nil, TrackMeta{}, err
		}
		defer response.Body.Close()
		img, err := png.Decode(response.Body)
		if err != nil {
			zaplog.ErrorC(ctx, "failed to decode cover art", zap.Error(err))
			return nil, TrackMeta{}, err
		}
		tag.SetCoverArt(&img)
	}
	output := new(bytes.Buffer)
	if err := tag.Save(output); err != nil {
		zaplog.ErrorC(ctx, "failed to save tag", zap.Error(err))
		return nil, TrackMeta{}, err
	}
	return output.Bytes(), bestMeta, nil
}

func (s *Service) GetYTMetaFromID(ctx context.Context, trackData track_sql.Track) (TrackMeta, error) {
	req, err := s.HTTPClient.CreateRequest(http.MethodGet, fmt.Sprintf("%s:%d/%s?id=%s", s.Config.BaseURL, s.Config.Ports.MusicAPI, s.Config.Endpoints.Meta, trackData.ID), nil)
	if err != nil {
		zaplog.ErrorC(ctx, "failed to create meta request", zap.Error(err))
		return TrackMeta{}, err
	}
	res, status, err := s.HTTPClient.DoRequest(req)
	if err != nil || status != http.StatusOK {
		zaplog.ErrorC(ctx, "error while sending meta request", zap.Error(err))
		return TrackMeta{}, err
	}
	var meta TrackMeta
	if err = json.Unmarshal(res, &meta); err != nil {
		zaplog.ErrorC(ctx, "failed to unmarshal meta response", zap.Error(err))
		return TrackMeta{}, err
	}
	return meta, nil
}

func (s *Service) GetSpotifyMeta(ctx context.Context, trackMeta TrackMeta) ([]TrackMeta, error) {
	searchTerm := fmt.Sprintf("track:%s artist:%s", trackMeta.Title, trackMeta.Artist)

	token, err := s.GetSpotifyToken(ctx)
	if err != nil {
		zaplog.ErrorC(ctx, "failed to get spotify token", zap.Error(err))
		return nil, err
	}

	authClient := spotifyauth.New().Client(ctx, token)
	spotifyClient := spotify.New(authClient)

	res, err := spotifyClient.Search(ctx, searchTerm, spotify.SearchTypeTrack)
	if err != nil {
		zaplog.ErrorC(ctx, "failed to search spotify", zap.Error(err))
		return nil, err
	}

	uniqueTrackMetas := make(map[spotify.ID]TrackMeta)
	for _, track := range res.Tracks.Tracks {
		resMeta := TrackMeta{}
		if len(track.Album.Images) > 0 {
			resMeta.CoverArtURL = track.Album.Images[0].URL
		}

		artists := make([]string, 0)
		for _, artist := range track.Artists {
			artists = append(artists, artist.Name)
		}

		resMeta.Artist = strings.Join(artists, ", ")
		resMeta.Album = track.Album.Name
		resMeta.Title = track.Name
		uniqueTrackMetas[track.ID] = resMeta
	}

	trackMetas := make([]TrackMeta, 0)
	for _, trackMeta := range uniqueTrackMetas {
		trackMetas = append(trackMetas, trackMeta)
	}

	return trackMetas, nil
}
func (s *Service) GetSpotifyToken(ctx context.Context) (*oauth2.Token, error) {
	token, err := s.SpotifyConfig.Token(ctx)
	if err != nil {
		zaplog.ErrorC(ctx, "failed to get spotify token", zap.Error(err))
		return nil, err
	}
	return token, nil
}

func (s *Service) GetBestMetaMatch(ctx context.Context, trackMeta TrackMeta, spotifyMetas []TrackMeta) TrackMeta {
	if len(spotifyMetas) == 0 {
		return trackMeta
	}
	coverArtist := s.CoverArtistCheck(ctx, trackMeta.Title)
	sanitizedTitle := s.SanitizeString(s.SanitizeParenthesis(trackMeta.Title))
	featStrippedTitle := strings.Split(sanitizedTitle, "feat")[0]
	titles := []string{trackMeta.Title, sanitizedTitle, featStrippedTitle}
	artists := []string{trackMeta.Artist}
	if coverArtist != "" {
		artists = append(artists, s.SanitizeAuthor(coverArtist))
	}
	sanitizedSplits := strings.Split(strings.ReplaceAll(sanitizedTitle, ":", "-"), "-")
	if len(sanitizedSplits) < 2 {
		titles = append(titles, sanitizedSplits[0])
	}
	if len(sanitizedSplits) == 2 {
		titles = append(titles, sanitizedSplits[0], sanitizedSplits[1])
		artists = append(artists, s.SanitizeAuthor(sanitizedSplits[0]), s.SanitizeAuthor(sanitizedSplits[1]))
	} else if len(sanitizedSplits) == 3 {
		titles = append(titles, sanitizedSplits[0], sanitizedSplits[1], sanitizedSplits[2], fmt.Sprintf("%s %s", sanitizedSplits[0], sanitizedSplits[1]), fmt.Sprintf("%s %s", sanitizedSplits[1], sanitizedSplits[2]))
		artists = append(artists, s.SanitizeAuthor(sanitizedSplits[0]), s.SanitizeAuthor(sanitizedSplits[1]), s.SanitizeAuthor(sanitizedSplits[2]), s.SanitizeAuthor(fmt.Sprintf("%s %s", sanitizedSplits[0], sanitizedSplits[1])), s.SanitizeAuthor(fmt.Sprintf("%s %s", sanitizedSplits[1], sanitizedSplits[2])))
	} else if len(sanitizedSplits) == 4 {
		titles = append(titles, sanitizedSplits[0], sanitizedSplits[1], sanitizedSplits[2], sanitizedSplits[3], fmt.Sprintf("%s %s", sanitizedSplits[0], sanitizedSplits[1]), fmt.Sprintf("%s %s", sanitizedSplits[1], sanitizedSplits[2]), fmt.Sprintf("%s %s", sanitizedSplits[2], sanitizedSplits[3]), fmt.Sprintf("%s %s %s", sanitizedSplits[0], sanitizedSplits[1], sanitizedSplits[2]), fmt.Sprintf("%s %s %s", sanitizedSplits[1], sanitizedSplits[2], sanitizedSplits[3]), fmt.Sprintf("%s %s", sanitizedSplits[0], sanitizedSplits[1]))
		artists = append(artists, s.SanitizeAuthor(sanitizedSplits[0]), s.SanitizeAuthor(sanitizedSplits[1]), s.SanitizeAuthor(sanitizedSplits[2]), s.SanitizeAuthor(sanitizedSplits[3]), s.SanitizeAuthor(fmt.Sprintf("%s %s", sanitizedSplits[0], sanitizedSplits[1])), s.SanitizeAuthor(fmt.Sprintf("%s %s", sanitizedSplits[1], sanitizedSplits[2])), s.SanitizeAuthor(fmt.Sprintf("%s %s", sanitizedSplits[2], sanitizedSplits[3])), s.SanitizeAuthor(fmt.Sprintf("%s %s %s", sanitizedSplits[0], sanitizedSplits[1], sanitizedSplits[2])), s.SanitizeAuthor(fmt.Sprintf("%s %s %s", sanitizedSplits[1], sanitizedSplits[2], sanitizedSplits[3])), s.SanitizeAuthor(fmt.Sprintf("%s %s", sanitizedSplits[0], sanitizedSplits[1])))
	}
	featStrippedSplits := strings.Split(strings.ReplaceAll(featStrippedTitle, ":", "-"), "-")
	if len(featStrippedSplits) < 2 {
		titles = append(titles, featStrippedSplits[0])
	}
	if len(featStrippedSplits) == 2 {
		titles = append(titles, featStrippedSplits[0], featStrippedSplits[1])
		artists = append(artists, s.SanitizeAuthor(featStrippedSplits[0]), s.SanitizeAuthor(featStrippedSplits[1]))
	} else if len(featStrippedSplits) == 3 {
		titles = append(titles, featStrippedSplits[0], featStrippedSplits[1], featStrippedSplits[2], fmt.Sprintf("%s %s", featStrippedSplits[0], featStrippedSplits[1]), fmt.Sprintf("%s %s", featStrippedSplits[1], featStrippedSplits[2]))
		artists = append(artists, s.SanitizeAuthor(featStrippedSplits[0]), s.SanitizeAuthor(featStrippedSplits[1]), s.SanitizeAuthor(featStrippedSplits[2]), s.SanitizeAuthor(fmt.Sprintf("%s %s", featStrippedSplits[0], featStrippedSplits[1])), s.SanitizeAuthor(fmt.Sprintf("%s %s", featStrippedSplits[1], featStrippedSplits[2])))
	} else if len(featStrippedSplits) == 4 {
		titles = append(titles, featStrippedSplits[0], featStrippedSplits[1], featStrippedSplits[2], featStrippedSplits[3], fmt.Sprintf("%s %s", featStrippedSplits[0], featStrippedSplits[1]), fmt.Sprintf("%s %s", featStrippedSplits[1], featStrippedSplits[2]), fmt.Sprintf("%s %s", featStrippedSplits[2], featStrippedSplits[3]), fmt.Sprintf("%s %s %s", featStrippedSplits[0], featStrippedSplits[1], featStrippedSplits[2]), fmt.Sprintf("%s %s %s", featStrippedSplits[1], featStrippedSplits[2], featStrippedSplits[3]), fmt.Sprintf("%s %s", featStrippedSplits[0], featStrippedSplits[1]))
		artists = append(artists, s.SanitizeAuthor(featStrippedSplits[0]), s.SanitizeAuthor(featStrippedSplits[1]), s.SanitizeAuthor(featStrippedSplits[2]), s.SanitizeAuthor(featStrippedSplits[3]), s.SanitizeAuthor(fmt.Sprintf("%s %s", featStrippedSplits[0], featStrippedSplits[1])), s.SanitizeAuthor(fmt.Sprintf("%s %s", featStrippedSplits[1], featStrippedSplits[2])), s.SanitizeAuthor(fmt.Sprintf("%s %s", featStrippedSplits[2], featStrippedSplits[3])), s.SanitizeAuthor(fmt.Sprintf("%s %s %s", featStrippedSplits[0], featStrippedSplits[1], featStrippedSplits[2])), s.SanitizeAuthor(fmt.Sprintf("%s %s %s", featStrippedSplits[1], featStrippedSplits[2], featStrippedSplits[3])), s.SanitizeAuthor(fmt.Sprintf("%s %s", featStrippedSplits[0], featStrippedSplits[1])))
	}
	for i, title := range titles {
		titles[i] = strings.Trim(strings.ReplaceAll(title, "  ", " "), " ")
	}
	for i, artist := range artists {
		artists[i] = strings.Trim(strings.ReplaceAll(artist, "  ", " "), " ")
	}

	for _, spotifyMeta := range spotifyMetas {
		if coverArtist != "" {
			if s.EqualIgnoringWhitespace(coverArtist, spotifyMeta.Artist) {
				for _, title := range titles {
					if s.EqualIgnoringWhitespace(title, spotifyMeta.Title) {
						return TrackMeta{Title: spotifyMeta.Title, Artist: spotifyMeta.Artist, Album: spotifyMeta.Album, Genre: trackMeta.Genre, CoverArtURL: spotifyMeta.CoverArtURL}
					}
				}
			}
		}
		for _, title := range titles {
			if s.EqualIgnoringWhitespace(title, spotifyMeta.Title) {
				for _, artist := range artists {
					if s.EqualIgnoringWhitespace(artist, spotifyMeta.Artist) {
						return TrackMeta{Title: spotifyMeta.Title, Artist: spotifyMeta.Artist, Album: spotifyMeta.Album, Genre: trackMeta.Genre, CoverArtURL: spotifyMeta.CoverArtURL}
					}
				}
			}
		}
	}

	return trackMeta
}

func (s *Service) SanitizeString(str string) string {
	regex := regexp.MustCompile(`[^a-zA-Z0-9\s\:\-]`)
	return regex.ReplaceAllString(str, "")
}

func (s *Service) SanitizeParenthesis(str string) string {
	regex := regexp.MustCompile(`\([^\(\)]*\)|\[[^\[\]]*\]`)
	return regex.ReplaceAllString(str, "")
}

func (s *Service) EqualIgnoringWhitespace(s1, s2 string) bool {
	// Remove all whitespace from both strings
	regex := regexp.MustCompile(`\s+`)
	cleanS1 := regex.ReplaceAllString(s1, "")
	cleanS2 := regex.ReplaceAllString(s2, "")

	// Compare the cleaned strings
	return strings.EqualFold(cleanS1, cleanS2)
}

func (s *Service) CoverArtistCheck(ctx context.Context, str string) string {
	str = strings.ToLower(str)
	parenthesisReg := regexp.MustCompile(`\([^\(\)]*\)|\[[^\[\]]*\]`)
	inParenthesis := parenthesisReg.FindAllString(str, -1)
	if len(inParenthesis) > 0 {
		for _, inParenthesisStr := range inParenthesis {
			if strings.Contains(strings.Trim(inParenthesisStr, " "), "cover by") {
				return strings.Trim(strings.Replace(inParenthesisStr, "cover by", "", -1), " ")
			} else if strings.Contains(strings.Trim(inParenthesisStr, " "), "covered by") {
				return strings.Trim(strings.Replace(inParenthesisStr, "covered by", "", -1), " ")
			} else if strings.HasSuffix(strings.Trim(inParenthesisStr, " "), "cover") {
				return strings.Trim(strings.Replace(inParenthesisStr, "cover", "", -1), " ")
			}
		}
	}
	return ""
}

func (s *Service) SanitizeAuthor(author string) string {
	author = strings.ToLower(author)
	r := regexp.MustCompile(` - official|-official|official| - vevo|-vevo|vevo|@| - topic|-topic|topic`)
	author = r.ReplaceAllString(author, "")
	author = strings.Trim(author, " ")
	return author
}
