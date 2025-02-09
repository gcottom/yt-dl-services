package track_sql

import (
	"database/sql"

	"github.com/gcottom/semaphore"
	"github.com/gcottom/yt-dl-services/downloader/config"
)

type Client struct {
	Config    *config.Config
	SQLClient *sql.DB
	Semaphore *semaphore.Semaphore
}

type Track struct {
	ID           string
	Title        string
	Author       string
	Artist       string
	Album        string
	Done         int
	Genre        string
	Error        int
	ErrorMessage string
}
