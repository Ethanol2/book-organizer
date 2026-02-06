package database

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
)

type Download struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	DirName   string    `json:"directory_name"`
	Files     BookFiles `json:"files"`
}

//#region Setters

func (c Client) AddDownload(tx *sql.Tx, dirPath string, files BookFiles) (*Download, error) {
	var err error

	indyTx := tx == nil
	if indyTx {
		tx, err = c.db.Begin()
		if err != nil {
			return nil, err
		}
		defer tx.Rollback()
	}

	id := uuid.New()

	audio, text, cover, err := files.ToJson()
	if err != nil {
		return nil, err
	}

	query := `
	INSERT INTO downloads
		(id, dir_name, audio_files, text_files, cover, created_at)
	VALUES
		(?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`
	_, err = tx.Exec(query, id, dirPath, audio, text, cover)
	if err != nil {
		return nil, err
	}

	if indyTx {
		err = tx.Commit()
		if err != nil {
			return nil, err
		}
	}

	log.Println("Added \"", dirPath, "\" to downloads")

	return c.GetDownload(id)
}

func (c Client) AddDownloads(downloads map[string]BookFiles) error {

	tx, err := c.db.Begin()
	if err != nil {
		return err
	}

	for name := range downloads {
		_, err = c.AddDownload(tx, name, downloads[name])
		if err != nil {
			return err
		}
	}

	tx.Commit()

	return nil
}

func (c Client) UpdateDownloadFiles(tx *sql.Tx, id uuid.UUID, files BookFiles) error {
	var err error
	indyTx := tx == nil
	if indyTx {
		tx, err = c.db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}

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
	_, err = tx.Exec(query, audio, text, cover, id)
	if err != nil {
		return err
	}

	if indyTx {
		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c Client) UpdateDownloadsFiles(files map[uuid.UUID]BookFiles) error {

	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for id := range files {
		err = c.UpdateDownloadFiles(tx, id, files[id])
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
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

func (c Client) GetDownloads() ([]Download, error) {

	rows, err := c.db.Query("SELECT * FROM downloads")
	if err != nil {
		return []Download{}, err
	}

	var downloads []Download

	for rows.Next() {
		var download Download
		var idStr string
		var audioJson string
		var textJson string

		err := rows.Scan(&idStr, &download.DirName, &audioJson, &textJson, &download.Files.Cover, &download.CreatedAt)
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

		err = download.Files.ParseAudioJson(audioJson)
		if err != nil {
			return nil, err
		}

		err = download.Files.ParseTextJson(textJson)
		if err != nil {
			return nil, err
		}

		downloads = append(downloads, download)
	}

	return downloads, nil
}

func (c Client) GetDownloadDir(id uuid.UUID) (string, error) {
	var dir string
	err := c.db.QueryRow("SELECT dir_name FROM downloads WHERE id = ?", id).Scan(&dir)
	if err != nil {
		return "", err
	}

	return dir, nil
}

//#region Helpers

func (c Client) getDownloadWithQuery(query string, args ...any) (*Download, error) {

	var download Download
	var idStr string
	var audioJson string
	var textJson string

	err := c.db.QueryRow(query, args...).Scan(&idStr, &download.DirName, &audioJson, &textJson, &download.Files.Cover, &download.CreatedAt)
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

	download.Files.Directory = &download.DirName

	err = download.Files.ParseAudioJson(audioJson)
	if err != nil {
		return nil, err
	}

	err = download.Files.ParseTextJson(textJson)
	if err != nil {
		return nil, err
	}

	return &download, err

}
