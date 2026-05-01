package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"path"
	"strconv"
	"strings"

	"github.com/Ethanol2/book-organizer/internal/fileManagement"
)

type Client struct {
	db *sql.DB
	tx *sql.Tx
}

func NewClient(dbPath string) (Client, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return Client{}, err
	}
	c := Client{db: db}
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
		has_metadata BOOLEAN NOT NULL,
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
		subtitle TEXT,
		publish_year INTEGER,
		description TEXT,
		tags TEXT,
		isbn TEXT UNIQUE,
		asin TEXT UNIQUE,
		publisher TEXT,
		directory TEXT,
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
			rank INTEGER NOT NULL,
			PRIMARY KEY (book_id, series_id),
			FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
			FOREIGN KEY (series_id) REFERENCES series(id) ON DELETE CASCADE
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
			rank INTEGER NOT NULL,
			PRIMARY KEY (%s_id, %s_id),
			FOREIGN KEY (%s_id) REFERENCES %s(id) ON DELETE CASCADE,
			FOREIGN KEY (%s_id) REFERENCES %s(id) ON DELETE CASCADE
		);
		`,
		type1Table, type2Table, type1, type2, type1, type2, type1, type1Table, type2, type2Table,
	)
	_, err := c.db.Exec(table)
	return err
}

func (c Client) InsertTestData(metadataPath, coversPath string) ([]Book, error) {

	fmt.Println()
	fmt.Print("======= Inserting Test Data =======")
	fmt.Println()

	var testBooks []BookParams
	err := json.Unmarshal([]byte(testData), &testBooks)
	if err != nil {
		return []Book{}, err
	}

	err = fileManagement.CreateDirectory(coversPath)
	if err != nil {
		fmt.Println("Error while creating the test covers directory \"", coversPath, "\"")
		return []Book{}, err
	}

	insertedBooks := []Book{}
	for _, params := range testBooks {
		book, err := c.AddBook(params)
		if err != nil {
			return []Book{}, err
		}

		insertedBooks = append(insertedBooks, book)

		if params.ISBN != nil {
			err = fileManagement.HandleTestCover(
				path.Join(coversPath, *params.ISBN+".jpg"),
				path.Join(metadataPath, book.Id.String()+".jpg"),
				fmt.Sprintf("https://covers.openlibrary.org/b/isbn/%s-L.jpg", *params.ISBN),
			)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	fmt.Println()
	fmt.Print("======= Finished Inserting Test Data =======")
	fmt.Println()

	return insertedBooks, nil
}

func (c *Client) Begin() error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	c.tx = tx
	return nil
}

func (c *Client) Rollback() error {
	if c.tx == nil {
		return nil
	}
	err := c.tx.Rollback()
	if err != nil {
		return err
	}
	c.tx = nil
	return nil
}

func (c *Client) Commit() error {
	if c.tx == nil {
		return nil
	}
	err := c.tx.Commit()
	if err != nil {
		return err
	}
	c.tx = nil
	return nil
}

func buildSearchQuery(filters map[string][]string) string {
	var advSearchFields = [...]string{"authors", "narrators", "genres", "series", "publisher", "publish_year", "isbn", "asin", "tags"}
	joinList := map[CategoryType]bool{
		Authors:   false,
		Genres:    false,
		Narrators: false,
		Series:    false,
	}

	cleanString := func(term string) string {
		term = strings.TrimSpace(term)
		term = strings.ReplaceAll(term, ",", "','")
		term = strings.ReplaceAll(term, "'", "''")
		return term
	}

	hasFilter := false
	filter := ""
	if search, ok := filters["search"]; ok {
		hasFilter = true
		term := cleanString(search[0])
		filter += `
		(books.title LIKE '%` + term + `%' OR 
		books.subtitle LIKE '%` + term + `%' OR 
		books.description LIKE '%` + term + `%') `
	}

	advFilter := []string{}
	for _, field := range advSearchFields {
		if terms, ok := filters[field]; ok {
			hasFilter = true

			if cat := stringToCategoryType(field); cat == NoType {
				// Field is a part of the books table

				split := strings.Split(terms[0], ",")
				for _, term := range split {
					term = strings.TrimSpace(term)
					advFilter = append(advFilter, `books.`+field+` LIKE "%`+term+`%" `)
				}
			} else {
				// Field is attached via joining table

				joinList[cat] = true
				advFilter = append(advFilter, string(cat)+".name IN ('"+cleanString(terms[0])+"')")
			}
		}
	}

	if files, ok := filters["files"]; ok {
		switch files[0] {
		case "with_files":
			hasFilter = true
			advFilter = append(advFilter, "directory IS NOT NULL")
		case "without_files":
			hasFilter = true
			advFilter = append(advFilter, "directory IS NULL")
		}
	}

	if hasFilter {
		if len(advFilter) > 0 && filter != "" {
			filter += " AND "
		}
		filter = "WHERE " + filter
	}

	sort := ""
	if sortType, ok := filters["sortBy"]; ok {
		order := ""
		if o, ok := filters["sortOrder"]; ok {
			order = o[0]
		}

		switch sortType[0] {
		case "title", "publisher", "created_at", "publish_year":
			sort = " ORDER BY " + sortType[0] + " " + order

		default:
			cat := stringToCategoryType(sortType[0])
			joinList[cat] = true
			sort = " ORDER BY " + string(cat) + ".name " + order

			if cat == Series {
				sort += ", books_series.series_index " + order
			}
		}
	}

	join := ""
	for cat, ok := range joinList {
		if ok {
			join += fmt.Sprintf(
				`JOIN books_%s ON books.id = books_%s.book_id JOIN %s ON books_%s.%s_id = %s.id `,

				string(cat),
				string(cat),
				string(cat),
				string(cat),
				categorySingular[cat],
				string(cat),
			)
		}
	}

	//fmt.Println(join + filter + strings.Join(advFilter, " AND ") + sort)

	return join + filter + strings.Join(advFilter, " AND ") + sort
}
func buildPageQuery(filters map[string][]string) (int, int, string) {

	countStr, hasCount := filters["count"]
	pageStr, hasPage := filters["page"]

	if !hasCount && !hasPage {
		return 20, 1, ""
	}

	var err error

	page := 1
	if hasPage {
		page, err = strconv.Atoi(pageStr[0])
		if err != nil {
			log.Println("Failed to convert page number to int =>", err)
			page = 1
		}
	}

	count := 20
	if hasCount {
		count, err = strconv.Atoi(countStr[0])
		if err != nil {
			log.Println("Failed to convert page count to int =>", err)
			count = 20
		}
	}

	return count, page, fmt.Sprintf(" LIMIT %d OFFSET %d", count, (page-1)*count)

}
