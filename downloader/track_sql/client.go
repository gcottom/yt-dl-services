package track_sql

import (
	"database/sql"

	"github.com/gcottom/semaphore"
	"github.com/gcottom/yt-dl-services/downloader/config"
	_ "modernc.org/sqlite"
)

func NewClient(cfg *config.Config) (*Client, error) {
	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		return nil, err
	}
	if err := CreateTables(db); err != nil {
		return nil, err
	}
	return &Client{
		Config:    cfg,
		SQLClient: db,
		Semaphore: semaphore.NewSemaphore(3),
	}, nil
}

func CreateTables(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS track (
		"id" TEXT NOT NULL PRIMARY KEY,
		"title" TEXT NOT NULL,
		"author" TEXT,
		"artist" TEXT,
		"album" TEXT,
		"done" INTEGER,
		"genre" TEXT,
		"error" INTEGER,
		"error_message" TEXT
	);`)
	return err
}
