package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
)

type Download struct {
	Id         uuid.UUID `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	Title      string    `json:"title"`
	AudioFiles FileList  `json:"audio_files"`
	TextFiles  FileList  `json:"text_files"`
	Cover      string    `json:"cover"`
	DirName    string    `json:"directory_name"`
}

func (c Client) CreateDownload(dirPath, title, cover string, audioFiles, textFiles []string) (*Download, error) {

	id := uuid.New()

	audioBytes, err := json.Marshal(FileList{audioFiles})
	if err != nil {
		return nil, err
	}
	textBytes, err := json.Marshal(FileList{textFiles})
	if err != nil {
		return nil, err
	}

	query := `
	INSERT INTO downloads
		(id, title, dir_name, audio_files, text_files, cover, created_at)
	VALUES
		(?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`
	_, err = c.db.Exec(query, id.String(), title, dirPath, string(audioBytes), string(textBytes), cover)
	if err != nil {
		return nil, err
	}

	log.Println("Added \"", title, "\" to downloads")

	return c.GetDownload(id)
}

func (c Client) GetDownload(id uuid.UUID) (*Download, error) {

	query := `
		SELECT * FROM downloads WHERE id = ?;	
	`

	return c.getDownloadWithQuery(query, id.String())
}

func (c Client) GetDownloadByDirectory(dir string) (*Download, error) {

	query := `
		SELECT * FROM downloads WHERE dir_name = ?;	
	`

	return c.getDownloadWithQuery(query, dir)

}

func (c Client) getDownloadWithQuery(query string, args ...any) (*Download, error) {

	var download Download
	var idStr string
	var audioJson string
	var textJson string

	err := c.db.QueryRow(query, args...).Scan(&idStr, &download.Title, &download.DirName, &audioJson, &textJson, &download.Cover, &download.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	download.Id, err = uuid.Parse(idStr)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(audioJson), &download.AudioFiles)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(textJson), &download.TextFiles)
	if err != nil {
		return nil, err
	}

	return &download, err

}
