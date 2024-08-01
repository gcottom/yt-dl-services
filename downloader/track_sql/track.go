package track_sql

import "context"

func (c *Client) InsertTrack(ctx context.Context, track Track) error {
	_, err := c.SQLClient.Exec("INSERT INTO track (id, title, author, artist, album, done, genre, error, error_message) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", track.ID, track.Title, track.Author, track.Artist, track.Album, track.Done, track.Genre, track.Error, track.ErrorMessage)
	return err
}

func (c *Client) GetTrack(ctx context.Context, id string) (Track, error) {
	row := c.SQLClient.QueryRow("SELECT id, title, author, artist, album, genre, error, error_message FROM track WHERE id = ?", id)
	var track Track
	err := row.Scan(&track.ID, &track.Title, &track.Author, &track.Artist, &track.Album, &track.Genre, &track.Error, &track.ErrorMessage)
	return track, err
}

func (c *Client) UpdateTrack(ctx context.Context, track Track) error {
	_, err := c.SQLClient.Exec("UPDATE track SET title = ?, author = ?, artist = ?, album = ?, done = ?, genre = ?, error = ?, error_message = ? WHERE id = ?", track.Title, track.Author, track.Artist, track.Album, track.Done, track.Genre, track.Error, track.ErrorMessage, track.ID)
	return err
}

func (c *Client) DeleteTrack(ctx context.Context, id string) error {
	_, err := c.SQLClient.Exec("DELETE FROM track WHERE id = ?", id)
	return err
}
