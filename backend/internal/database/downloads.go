package database

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/Ethanol2/book-organizer/internal/fileManagement"

	"github.com/google/uuid"
)

type Download struct {
	Id        uuid.UUID            `json:"id"`
	CreatedAt time.Time            `json:"created_at"`
	Files     fileManagement.Files `json:"files"`
}

//#region Setters

func (c *Client) AddDownload(files fileManagement.Files) error {
	var err error

	id := uuid.New()

	audio, text, err := files.FileListsToJson()
	if err != nil {
		return err
	}

	query := `
	INSERT INTO downloads
		(id, dir_name, audio_files, text_files, cover, has_metadata, created_at)
	VALUES
		(?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`
	_, err = c.handler.Exec(query, id, files.Root, audio, text, files.Cover, files.HasMetadata)
	if err != nil {
		return err
	}

	log.Println("Added \"", *files.Root, "\" to downloads")

	return nil
}

// Handles the transaction internallly
func (c *Client) AddDownloads(downloads []fileManagement.Files) error {

	return c.HandleTransaction(func(c *Client) error {
		for name := range downloads {
			err := c.AddDownload(downloads[name])
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// Handles the transaction internally
func (c *Client) UpdateDownloadFiles(id uuid.UUID, files fileManagement.Files) error {

	return c.HandleTransaction(func(c *Client) error {
		audio, text, err := files.FileListsToJson()
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
		_, err = c.handler.Exec(query, audio, text, files.Cover, id)
		if err != nil {
			return err
		}

		return nil
	})

}

// Handles the transaction internally
func (c *Client) DeleteDownload(id uuid.UUID) error {

	return c.HandleTransaction(func(c *Client) error {
		_, err := c.handler.Exec("DELETE FROM downloads WHERE id = ?", id)
		if err != nil {
			return err
		}
		return nil
	})
}

//#region Getters

func (c *Client) GetDownload(id uuid.UUID) (*Download, error) {

	query := `
	SELECT * FROM downloads WHERE id = ?;	
	`

	return c.getDownloadWithQuery(query, id.String())
}

func (c *Client) GetDownloadByDirectory(dir string) (*Download, error) {

	query := `
		SELECT * FROM downloads WHERE dir_name = ?;	
	`

	return c.getDownloadWithQuery(query, dir)

}

// Handles the transaction internally
func (c *Client) GetAllDownloadsIdsAndDirs() ([]uuid.UUID, []string, error) {

	var ids []uuid.UUID
	var dirs []string
	err := c.HandleTransaction(func(c *Client) error {
		query := `
		SELECT id, dir_name FROM downloads
	`

		rows, err := c.handler.Query(query)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var idStr string
			var dir string

			if err := rows.Scan(&idStr, &dir); err != nil {
				return err
			}

			id, err := uuid.Parse(idStr)
			if err != nil {
				log.Println(err)
				continue
			}

			ids = append(ids, id)
			dirs = append(dirs, dir)
		}
		return nil
	})
	if err != nil {
		return []uuid.UUID{}, []string{}, err
	}

	return ids, dirs, nil
}

func (c *Client) GetDownloads() ([]Download, error) {

	rows, err := c.handler.Query("SELECT * FROM downloads")
	if err != nil {
		return []Download{}, err
	}

	var downloads []Download

	for rows.Next() {
		var download Download
		var idStr string
		var audioJson string
		var textJson string

		err := rows.Scan(&idStr, &download.Files.Root, &audioJson, &textJson, &download.Files.Cover, &download.Files.HasMetadata, &download.CreatedAt)
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

func (c *Client) GetDownloadDir(id uuid.UUID) (string, error) {
	var dir string
	err := c.handler.QueryRow("SELECT dir_name FROM downloads WHERE id = ?", id).Scan(&dir)
	if err != nil {
		return "", err
	}

	return dir, nil
}

//#region Helpers

func (c *Client) getDownloadWithQuery(query string, args ...any) (*Download, error) {

	var download Download
	var idStr string
	var audioJson string
	var textJson string

	err := c.handler.QueryRow(query, args...).Scan(&idStr, &download.Files.Root, &audioJson, &textJson, &download.Files.Cover, &download.Files.HasMetadata, &download.CreatedAt)
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

	return &download, err

}

func (c *Client) RemoveAllDownloads() error {

	_, err := c.handler.Exec("DELETE FROM downloads")
	if err != nil {
		return err
	}
	return nil
}
