package track_sql

import (
	"database/sql"

	"github.com/gcottom/yt-dl-services/downloader/config"
	_ "github.com/mattn/go-sqlite3"
)

func NewClient(cfg *config.Config) (*Client, error) {
	db, err := sql.Open("sqlite3", cfg.DBPath)
	if err != nil {
		return nil, err
	}
	if err := CreateTables(db); err != nil {
		return nil, err
	}
	return &Client{
		Config:    cfg,
		SQLClient: db,
	}, nil
}

func CreateTables(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS tracks (
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
