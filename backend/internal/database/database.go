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
		return Client{}, err
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
		description TEXT,
		tags TEXT,
		isbn TEXT UNIQUE,
		asin TEXT UNIQUE,
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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL	
	);	
	`
	_, err = client.db.Exec(authorsTable)
	if err != nil {
		return err
	}

	narratorsTable := `
	CREATE TABLE IF NOT EXISTS narrators (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL
	);
	`
	_, err = client.db.Exec(narratorsTable)
	if err != nil {
		return err
	}

	seriesTable := `
	CREATE TABLE IF NOT EXISTS series (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL
	);
	`
	_, err = client.db.Exec(seriesTable)
	if err != nil {
		return err
	}

	genresTable := `
	CREATE TABLE IF NOT EXISTS genres (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL
	);
	`
	_, err = client.db.Exec(genresTable)
	if err != nil {
		return err
	}

	// Joining Tables

	booksSeriestable :=
		`
		CREATE TABLE IF NOT EXISTS books_series (
			book_id TEXT NOT NULL,
			series_id TEXT NOT NULL,
			series_index TEXT,
			PRIMARY KEY (book_id, series_id),
			FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
			FOREIGN KEY (series_id) REFERENCES series(id)
		);
		`
	_, err = client.db.Exec(booksSeriestable)
	if err != nil {
		return err
	}

	err = client.generateJoiningTable("book", "books", categorySingular[Authors], string(Authors))
	if err != nil {
		return err
	}

	err = client.generateJoiningTable("book", "books", categorySingular[Narrators], string(Narrators))
	if err != nil {
		return err
	}

	err = client.generateJoiningTable("book", "books", categorySingular[Genres], string(Genres))
	if err != nil {
		return err
	}

	return nil
}

func (c Client) generateJoiningTable(type1, type1Table, type2, type2Table string) error {
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

func (c Client) InsertTestData() error {

	tx, _ := c.db.Begin()
	defer tx.Rollback()

	fmt.Print("\n======= Inserting Test Data =======\n\n")

	genres := []string{
		"Romance", "Comedy", "Drama",
	}

	authors := []string{
		"John Scalzi", "Andy Weir", "Shakespear", "Zogarth", "James S.A. Corey",
	}

	narrators := []string{
		"Wil Wheaton", "Heath Miller", "Eric Mock",
	}

	series := []string{
		"The Expanse", "Old Man's War", "The Primal Hunter",
	}

	createCategories := func(categoryType CategoryType, values []string) error {

		for _, value := range values {
			_, err := c.AddCategory(tx, categoryType, value)
			if err != nil {
				return err
			}
		}
		return nil
	}

	err := createCategories(Genres, genres)
	if err != nil {
		return err
	}

	err = createCategories(Authors, authors)
	if err != nil {
		return err
	}

	err = createCategories(Narrators, narrators)
	if err != nil {
		return err
	}

	err = createCategories(Series, series)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	intPtr := func(i int) *int {
		return &i
	}
	strPtr := func(s string) *string {
		return &s
	}

	// 1. The Martian by Andy Weir
	theMartian := CreateBookParams{
		Title:       "The Martian",
		Description: "A stranded astronaut must use his ingenuity to survive on Mars.",
		Year:        intPtr(2011),
		ISBN:        "9780553418026",
		Tags:        []string{"Sci-Fi", "Survival", "Space"},
		Publisher:   "Crown",
		Authors:     []Category{{Name: "Andy Weir"}},
		Genres:      []Category{{Name: "Sci-Fi"}},
		Narrators:   []Category{{Name: "R.C. Bray"}},
	}

	// 2. The Primal Hunter by Zogarth
	thePrimalHunter := CreateBookParams{
		Title:       "The Primal Hunter",
		Description: "A fast-paced LitRPG adventure where the world undergoes a tutorial.",
		Year:        intPtr(2022),
		ASIN:        "B09MTY98S8",
		Tags:        []string{"LitRPG", "Progression Fantasy", "Action"},
		Publisher:   "Aethon Books",
		Authors:     []Category{{Name: "Zogarth"}},
		Genres:      []Category{{Name: "Fantasy"}},
		Series:      []Category{{Name: "The Primal Hunter", Index: strPtr("1")}},
		Narrators:   []Category{{Name: "Travis Baldree"}},
	}

	_, err = c.AddBook(theMartian)
	if err != nil {
		return err
	}

	_, err = c.AddBook(thePrimalHunter)
	if err != nil {
		return err
	}

	fmt.Print("\n======= Finished Inserting Test Data =======\n\n")

	return nil
}
