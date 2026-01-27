package database

import (
	"database/sql"
	"encoding/json"
)

type Client struct {
	db *sql.DB
}
type BookFiles struct {
	AudioFiles FileList `json:"audio_files"`
	TextFiles  FileList `json:"text_files"`
	Cover      string   `json:"cover"`
}
type FileList struct {
	Files []string `json:"files"`
}

func NewClient(dbPath string) (Client, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return Client{}, err
	}
	c := Client{db}
	err = c.handleMigration()
	if err != nil {
		return Client{db}, err
	}
	return c, nil
}

func (client *Client) handleMigration() error {
	downloadsTable := `
	CREATE TABLE IF NOT EXISTS downloads (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		dir_name TEXT NOT NULL,
		audio_files TEXT,
		text_files TEXT,
		cover TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);	
	`
	_, err := client.db.Exec(downloadsTable)
	return err
}

func (files BookFiles) ToJson() (string, string, string, error) {

	audioBytes, err := json.Marshal(files.AudioFiles)
	if err != nil {
		return "", "", "", err
	}
	textBytes, err := json.Marshal(files.TextFiles)
	if err != nil {
		return "", "", "", err
	}

	return string(audioBytes), string(textBytes), files.Cover, nil
}
