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
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Title     string    `json:"title"`
	DirName   string    `json:"directory_name"`
	Files     BookFiles `json:"files"`
}

//#region Setters

func (c Client) CreateDownload(dirPath, title string, files BookFiles) (*Download, error) {

	id := uuid.New()

	audio, text, cover, err := files.ToJson()
	if err != nil {
		return nil, err
	}

	query := `
	INSERT INTO downloads
		(id, title, dir_name, audio_files, text_files, cover, created_at)
	VALUES
		(?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`
	_, err = c.db.Exec(query, id, title, dirPath, audio, text, cover)
	if err != nil {
		return nil, err
	}

	log.Println("Added \"", title, "\" to downloads")

	return c.GetDownload(id)
}

func (c Client) UpdateDownloadFiles(id uuid.UUID, files BookFiles) error {

	audio, text, cover, err := files.ToJson()
	if err != nil {
		return err
	}

	query := `
		UPDATE downloads
		SET
			audio_files = ?,
			text_files = ?,
			cover = ?
		WHERE id = ?
	`
	_, err = c.db.Exec(query, audio, text, cover, id)
	if err != nil {
		return err
	}

	return nil
}

func (c Client) DeleteDownload(id uuid.UUID) error {
	_, err := c.db.Exec("DELETE FROM downloads WHERE id = ?", id)
	if err != nil {
		return err
	}
	return nil
}

//#region Getters

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

func (c Client) GetAllDownloadsIdsAndDirs() ([]uuid.UUID, []string, error) {

	query := `
		SELECT id, dir_name FROM downloads
	`

	rows, err := c.db.Query(query)
	if err != nil {
		return []uuid.UUID{}, []string{}, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	var dirs []string

	for rows.Next() {
		var idStr string
		var dir string

		if err := rows.Scan(&idStr, &dir); err != nil {
			return []uuid.UUID{}, []string{}, err
		}

		id, err := uuid.Parse(idStr)
		if err != nil {
			log.Println(err)
			continue
		}

		ids = append(ids, id)
		dirs = append(dirs, dir)
	}

	return ids, dirs, nil
}

//#region Helpers

func (c Client) getDownloadWithQuery(query string, args ...any) (*Download, error) {

	var download Download
	var idStr string
	var audioJson string
	var textJson string

	err := c.db.QueryRow(query, args...).Scan(&idStr, &download.Title, &download.DirName, &audioJson, &textJson, &download.Files.Cover, &download.CreatedAt)
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

	err = json.Unmarshal([]byte(audioJson), &download.Files.AudioFiles)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(textJson), &download.Files.TextFiles)
	if err != nil {
		return nil, err
	}

	return &download, err

}

func (c Client) getDownloadFilesWithQuery(query string, args ...any) (uuid.UUID, *BookFiles, error) {

	var files BookFiles
	var idStr string
	var audioJson string
	var textJson string

	err := c.db.QueryRow(query, args...).Scan(&idStr, &audioJson, &textJson, &files.Cover)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, nil, nil
		}
		return uuid.Nil, nil, err
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, nil, err
	}

	err = json.Unmarshal([]byte(audioJson), &files.AudioFiles)
	if err != nil {
		return uuid.Nil, nil, err
	}

	err = json.Unmarshal([]byte(textJson), &files.TextFiles)
	if err != nil {
		return uuid.Nil, nil, err
	}

	return id, &files, err

}
