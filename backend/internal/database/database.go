package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

type Client struct {
	db *sql.DB
	tx *sql.Tx
}

func NewClient(dbPath string) (Client, error) {
	db, err := sql.Open("sqlite3", dbPath)
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

func (c Client) InsertTestData() error {

	c.Begin()
	defer c.Rollback()

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
			_, err := c.AddCategory(categoryType, value)
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

	err = c.tx.Commit()
	if err != nil {
		return err
	}

	// 1. The Martian by Andy Weir
	theMartianJson := `
	{
		"title": "The Martian",
		"description": "Six days ago, astronaut Mark Watney became one of the first people to walk on Mars. Now, he's sure he'll be the first person to die there.",
		"year": 2011,
		"isbn": "9780553418026",
		"asin": "B00EMXBDMA",
		"tags": ["Sci-Fi", "Survival", "Hard Science Fiction"],
		"publisher": "Crown",
		"series": [],
		"authors": [{"name": "Andy Weir"}],
		"genres": [{"name": "Science Fiction"}],
		"narrators": [{"name": "R.C. Bray"}],
		"cover": "https://ia800505.us.archive.org/view_archive.php?archive=/35/items/l_covers_0014/l_covers_0014_64.zip&file=0014641755-L.jpg"
	  }	  
	`
	var theMartian BookParams
	err = json.Unmarshal([]byte(theMartianJson), &theMartian)
	if err != nil {
		return err
	}

	// 2. The Primal Hunter by Zogarth
	thePrimalHunterJson := `
	{
		"title": "The Primal Hunter",
		"description": "A world changed. An ancient system awakened. Jake, a corporate drone, finds himself in a tutorial that will change his life forever.",
		"year": 2022,
		"isbn": "9798834943709",
		"asin": "B09MV5TTSM",
		"tags": ["LitRPG", "Progression Fantasy", "Action"],
		"publisher": "Aethon Books",
		"series": [
		  {
			"name": "The Primal Hunter",
			"index": "1"
		  }
		],
		"authors": [{"name": "Zogarth"}],
		"genres": [{"name": "Fantasy"}, {"name": "LitRPG"}],
		"narrators": [{"name": "Travis Baldree"}]
	  }	  
	`
	var thePrimalHunter BookParams
	err = json.Unmarshal([]byte(thePrimalHunterJson), &thePrimalHunter)
	if err != nil {
		return err
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

	hasFilter := false
	filter := ""
	if search, ok := filters["search"]; ok {
		hasFilter = true
		filter += `
		(books.title LIKE '%` + search[0] + `%' OR 
		books.subtitle LIKE '%` + search[0] + `%' OR 
		books.description LIKE '%` + search[0] + `%') `
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
				advFilter = append(advFilter, string(cat)+".name IN ('"+strings.ReplaceAll(terms[0], ",", "','")+"')")
			}
		}
	}

	if hasFilter {
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
			sort = " ORDER BY " + string(cat) + ".name"
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

	return join + filter + strings.Join(advFilter, " AND ") + sort
}
