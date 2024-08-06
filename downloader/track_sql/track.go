package track_sql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gcottom/retry"
)

func (c *Client) InsertTrack(ctx context.Context, track Track) error {
	_, err := c.GetTrack(ctx, track.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.Semaphore.Acquire()
			_, err := retry.Retry(retry.NewAlgFibonacciDefault(), 5, c.SQLClient.Exec, "INSERT INTO track (id, title, author, artist, album, done, genre, error, error_message) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", track.ID, track.Title, track.Author, track.Artist, track.Album, track.Done, track.Genre, track.Error, track.ErrorMessage)
			c.Semaphore.Release()
			return err
		} else {
			return err
		}
	}
	return c.UpdateTrack(ctx, track)
}

func (c *Client) GetTrack(ctx context.Context, id string) (Track, error) {
	c.Semaphore.Acquire()
	row := c.SQLClient.QueryRow("SELECT id, title, author, artist, album, done, genre, error, error_message FROM track WHERE id = ?", id)
	var track Track
	err := row.Scan(&track.ID, &track.Title, &track.Author, &track.Artist, &track.Album, &track.Done, &track.Genre, &track.Error, &track.ErrorMessage)
	c.Semaphore.Release()
	return track, err
}

func (c *Client) UpdateTrack(ctx context.Context, track Track) error {
	c.Semaphore.Acquire()
	_, err := retry.Retry(retry.NewAlgFibonacciDefault(), 5, c.SQLClient.Exec, "UPDATE track SET title = ?, author = ?, artist = ?, album = ?, done = ?, genre = ?, error = ?, error_message = ? WHERE id = ?", track.Title, track.Author, track.Artist, track.Album, track.Done, track.Genre, track.Error, track.ErrorMessage, track.ID)
	c.Semaphore.Release()
	return err
}

func (c *Client) DeleteTrack(ctx context.Context, id string) error {
	c.Semaphore.Acquire()
	_, err := retry.Retry(retry.NewAlgFibonacciDefault(), 5, c.SQLClient.Exec, "DELETE FROM track WHERE id = ?", id)
	c.Semaphore.Release()
	return err
}
