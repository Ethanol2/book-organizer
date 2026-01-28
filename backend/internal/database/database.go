package database

import (
	"database/sql"
	"fmt"
)

type Client struct {
	db *sql.DB
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
		dir_name TEXT NOT NULL,
		audio_files TEXT,
		text_files TEXT,
		cover TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);	
	`
	_, err := client.db.Exec(downloadsTable)
	if err != nil {
		return err
	}

	booksTable := `
	CREATE TABLE IF NOT EXISTS books (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		publish_year INTEGER,
		series_index INTEGER,
		description TEXT,
		tags TEXT,
		isbn TEXT,
		asin TEXT,
		publisher TEXT,
		audio_files TEXT,
		text_files TEXT,
		cover TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = client.db.Exec(booksTable)
	if err != nil {
		return err
	}

	authorsTable := `
	CREATE TABLE IF NOT EXISTS authors (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL	
	);	
	`
	_, err = client.db.Exec(authorsTable)
	if err != nil {
		return err
	}

	narratorsTable := `
	CREATE TABLE IF NOT EXISTS narrators (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL
	);
	`
	_, err = client.db.Exec(narratorsTable)
	if err != nil {
		return err
	}

	seriesTable := `
	CREATE TABLE IF NOT EXISTS series (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL
	);
	`
	_, err = client.db.Exec(seriesTable)
	if err != nil {
		return err
	}

	genresTable := `
	CREATE TABLE IF NOT EXISTS genres (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL
	);
	`
	_, err = client.db.Exec(genresTable)
	if err != nil {
		return err
	}

	// Joining Tables

	err = client.generateJoiningTableQuery("book", "books", "series", "series")
	if err != nil {
		return err
	}

	err = client.generateJoiningTableQuery("book", "books", "author", "authors")
	if err != nil {
		return err
	}

	err = client.generateJoiningTableQuery("book", "books", "narrator", "narrators")
	if err != nil {
		return err
	}

	err = client.generateJoiningTableQuery("book", "books", "genre", "genres")
	if err != nil {
		return err
	}

	return nil
}

func (c Client) generateJoiningTableQuery(type1, type1Table, type2, type2Table string) error {
	table := fmt.Sprintf(
		`
		CREATE TABLE IF NOT EXISTS %s_%s (
			%s_id TEXT NOT NULL,
			%s_id TEXT NOT NULL,
			PRIMARY KEY (%s_id, %s_id),
			FOREIGN KEY (%s_id) REFERENCES %s(id) ON DELETE CASCADE,
			FOREIGN KEY (%s_id) REFERENCES %s(id)
		);
		`,
		type1Table, type2Table, type1, type2, type1, type2, type1, type1Table, type2, type2Table,
	)
	_, err := c.db.Exec(table)
	return err
}
