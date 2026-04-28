package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Book struct {
	Id          *uuid.UUID `json:"id,omitempty"`
	Title       string     `json:"title"`
	Subtitle    *string    `json:"subtitle"`
	Description *string    `json:"description"`
	Year        *int       `json:"year"`
	ISBN        *string    `json:"isbn"`
	ASIN        *string    `json:"asin"`
	Tags        []string   `json:"tags"`
	Publisher   *string    `json:"publisher"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`

	// Categories
	Series    []Category `json:"series"`
	Authors   []Category `json:"authors"`
	Genres    []Category `json:"genres"`
	Narrators []Category `json:"narrators"`

	Files BookFiles `json:"files,omitempty"`
}

type BookOverview struct {
	Id       uuid.UUID  `json:"id"`
	Title    string     `json:"title"`
	Subtitle *string    `json:"subtitle"`
	Authors  []Category `json:"authors"`
	Cover    *string    `json:"cover"`
	HasFiles bool       `json:"has_files"`
}

type BookSearchResults[T []BookOverview | []Book] struct {
	ResultsCount int `json:"results_count"`
	Count        int `json:"count"`
	Page         int `json:"page"`
	Items        T   `json:"items,omitempty"`
}

type BookParams struct {
	Title       *string   `json:"title"`
	Subtitle    *string   `json:"subtitle"`
	Description *string   `json:"description"`
	Year        *int      `json:"year"`
	ISBN        *string   `json:"isbn"`
	ASIN        *string   `json:"asin"`
	Tags        *[]string `json:"tags"`
	Publisher   *string   `json:"publisher"`

	// Categories
	Series    *[]Category `json:"series"`
	Authors   *[]Category `json:"authors"`
	Genres    *[]Category `json:"genres"`
	Narrators *[]Category `json:"narrators"`

	// URIs
	Cover *string `json:"cover"`
	Key   *string `json:"key"`
}

func (c Client) CheckBookExists(id uuid.UUID) (bool, error) {
	var exists bool
	err := c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM books WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (c Client) AddBook(params BookParams) (Book, error) {

	err := c.Begin()
	if err != nil {
		return Book{}, err
	}
	defer c.Rollback()

	id := uuid.New()

	tagsJson, err := json.Marshal(params.Tags)
	if err != nil {
		return Book{}, err
	}

	if params.ISBN != nil && *params.ISBN == "" {
		params.ISBN = nil
	}
	if params.ASIN != nil && *params.ASIN == "" {
		params.ASIN = nil
	}

	query := `
	INSERT INTO books
		(id, title, subtitle, publish_year, description, tags, isbn, asin, publisher, cover, directory, audio_files, text_files, created_at, updated_at)
	VALUES
		(?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, NULL, NULL, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)	
	`

	_, err = c.tx.Exec(query, id, params.Title, params.Subtitle, params.Year, params.Description, string(tagsJson), params.ISBN, params.ASIN, params.Publisher)
	if err != nil {
		return Book{}, err
	}

	sortCats := func(catType CategoryType, catsPtr *[]Category) error {

		if catsPtr == nil {
			return nil
		}
		cats := *catsPtr

		log.Println("Associating", catType)

		for i, cat := range cats {
			cat.Type = catType
			err := c.associateBookAndCategoryType(id.String(), cat, i)
			if err != nil {
				return err
			}
		}
		return nil
	}

	err = sortCats(Series, params.Series)
	if err != nil {
		return Book{}, err
	}

	err = sortCats(Genres, params.Genres)
	if err != nil {
		return Book{}, err
	}

	err = sortCats(Narrators, params.Narrators)
	if err != nil {
		return Book{}, err
	}

	err = sortCats(Authors, params.Authors)
	if err != nil {
		return Book{}, err
	}

	err = c.Commit()
	if err != nil {
		return Book{}, err
	}
	log.Println("Added \"", *params.Title, "\" to books")

	return c.GetBook(id)
}

func (c Client) GetBook(id uuid.UUID) (Book, error) {

	var book Book
	var tagsStr *string
	var audioStr *string
	var textStr *string

	err := c.db.QueryRow("SELECT * FROM books WHERE id = ?", id).Scan(
		&book.Id,
		&book.Title,
		&book.Subtitle,
		&book.Year,
		&book.Description,
		&tagsStr,
		&book.ISBN,
		&book.ASIN,
		&book.Publisher,
		&book.Files.Root,
		&audioStr,
		&textStr,
		&book.Files.Cover,
		&book.CreatedAt,
		&book.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Book{}, nil
		}
		return Book{}, err
	}

	if tagsStr != nil {
		err = json.Unmarshal([]byte(*tagsStr), &book.Tags)
		if err != nil {
			return Book{}, err
		}
	}

	if audioStr != nil {
		err = book.Files.ParseAudioJson(*audioStr)
		if err != nil {
			return Book{}, err
		}
	}

	if textStr != nil {
		err = book.Files.ParseTextJson(*textStr)
		if err != nil {
			return Book{}, err
		}
	}

	err = book.getBookCategories(c)
	if err != nil {
		return Book{}, err
	}

	//defer log.Println("Retrieved \"", book.Title, "\" from books")

	return book, nil
}

func (c Client) GetBooks(filters map[string][]string) (BookSearchResults[[]Book], error) {

	err := c.Begin()
	if err != nil {
		return BookSearchResults[[]Book]{}, err
	}
	defer c.Rollback()

	books := []Book{}

	count, page, pageQuery := buildPageQuery(filters)

	query := "SELECT *, (SELECT COUNT(*) FROM books) AS total_count FROM books " + buildSearchQuery(filters) + pageQuery
	rows, err := c.tx.Query(query)
	if err != nil {
		log.Println("Query:\n", query)
		return BookSearchResults[[]Book]{}, err
	}
	defer rows.Close()

	totalCount := 0
	for rows.Next() {
		var book Book
		var tagsStr *string
		var audioStr *string
		var textStr *string

		err := rows.Scan(
			&book.Id,
			&book.Title,
			&book.Subtitle,
			&book.Year,
			&book.Description,
			&tagsStr,
			&book.ISBN,
			&book.ASIN,
			&book.Publisher,
			&book.Files.Root,
			&audioStr,
			&textStr,
			&book.Files.Cover,
			&book.CreatedAt,
			&book.UpdatedAt,
			&totalCount,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		if audioStr != nil {
			err = book.Files.ParseAudioJson(*audioStr)
			if err != nil {
				return BookSearchResults[[]Book]{}, err
			}
		}

		if textStr != nil {
			err = book.Files.ParseTextJson(*textStr)
			if err != nil {
				return BookSearchResults[[]Book]{}, err
			}
		}

		err = book.getBookCategories(c)
		if err != nil {
			log.Println(err)
			continue
		}

		books = append(books, book)
	}

	err = c.Commit()
	if err != nil {
		return BookSearchResults[[]Book]{}, err
	}

	return BookSearchResults[[]Book]{
		ResultsCount: totalCount,
		Count:        count,
		Page:         page,
		Items:        books,
	}, nil
}

func (c Client) GetBooksSummary(filters map[string][]string) (BookSearchResults[[]BookOverview], error) {

	err := c.Begin()
	if err != nil {
		return BookSearchResults[[]BookOverview]{}, err
	}
	defer c.Rollback()

	countLimit, page, pageQuery := buildPageQuery(filters)
	query := "SELECT books.id, books.title, books.subtitle, books.cover, books.directory, (SELECT COUNT(*) FROM books) AS total_count  FROM books " + buildSearchQuery(filters) + pageQuery

	rows, err := c.tx.Query(query)
	if err != nil {
		log.Println("Query:\n", query)
		return BookSearchResults[[]BookOverview]{}, err
	}
	defer rows.Close()

	totalCount := 0
	var books []BookOverview
	for rows.Next() {
		var book BookOverview
		var dir *string
		err = rows.Scan(&book.Id, &book.Title, &book.Subtitle, &book.Cover, &dir, &totalCount)
		if err != nil {
			return BookSearchResults[[]BookOverview]{}, err
		}

		book.Authors, err = c.GetCategoryTypesAssociatedWithBook(book.Id.String(), Authors)
		if err != nil {
			return BookSearchResults[[]BookOverview]{}, err
		}

		book.HasFiles = dir != nil

		books = append(books, book)
	}

	return BookSearchResults[[]BookOverview]{
		ResultsCount: totalCount,
		Count:        countLimit,
		Page:         page,
		Items:        books,
	}, nil
}

func (c Client) AssociateBookAndDownload(bookId, downloadId uuid.UUID, author, series, bookDir string) (Book, error) {

	err := c.Begin()
	if err != nil {
		return Book{}, err
	}
	defer c.Rollback()

	var files BookFiles
	var Audio *string
	var Text *string

	err = c.tx.QueryRow(`
	SELECT dir_name, audio_files, text_files, cover FROM downloads WHERE id = ?
	`, downloadId).Scan(&files.Root, &Audio, &Text, &files.Cover)
	if err != nil {
		return Book{}, err
	}

	err = files.ParseAudioJson(*Audio)
	if err != nil {
		return Book{}, err
	}
	err = files.ParseTextJson(*Text)
	if err != nil {
		return Book{}, err
	}

	files.UpdateDirectory(path.Join(author, series, bookDir))
	tmpBook := Book{Id: &bookId, Files: files}
	err = tmpBook.updateBookFiles(c.tx)
	if err != nil {
		return Book{}, err
	}

	err = c.Commit()
	if err != nil {
		return Book{}, err
	}

	return c.GetBook(bookId)
}

// Returns the updated book and a bool that says whether the file path has updated
func (c Client) UpdateBook(id uuid.UUID, update BookParams) (Book, bool, error) {

	indyTx := c.tx == nil
	if indyTx {
		err := c.Begin()
		if err != nil {
			return Book{}, false, err
		}
		defer c.Rollback()
	}

	setParts := []string{"updated_at = CURRENT_TIMESTAMP"}
	args := []interface{}{}
	needsFileUpdate := false

	add := func(part string, arg interface{}) {
		setParts = append(setParts, part+" = ?")
		args = append(args, arg)
	}

	if update.Title != nil {
		add("title", update.Title)
		needsFileUpdate = true
	}
	if update.Year != nil {
		add("publish_year", update.Year)
	}
	if update.Description != nil {
		add("description", update.Description)
	}
	if update.Tags != nil {
		tagsJson, err := json.Marshal(update.Tags)
		if err != nil {
			return Book{}, false, err
		}
		add("tags", string(tagsJson))
	}
	if update.ISBN != nil {
		add("isbn", update.ISBN)
	}
	if update.ASIN != nil {
		add("asin", update.ASIN)
	}
	if update.Publisher != nil {
		add("publisher", update.Publisher)
	}

	if len(setParts) > 0 {
		query := "UPDATE books SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
		args = append(args, id)
		_, err := c.tx.Exec(query, args...)
		if err != nil {
			return Book{}, false, err
		}
	}

	idStr := id.String()
	handleCategories := func(deleteQuery string, update, old []Category, catType CategoryType) error {
		new := []Category{}
		removed := old

		for _, cat := range update {
			if slices.ContainsFunc(old, func(c Category) bool {
				return c.Name == cat.Name && c.Index == cat.Index
			}) {
				index := slices.IndexFunc(removed, func(c Category) bool {
					return c.Name == cat.Name
				})
				removed = slices.Delete(removed, index, index+1)
			} else {
				cat.Type = catType
				new = append(new, cat)
			}
		}

		for _, cat := range removed {
			log.Println("Removing", cat.Name)
			_, err := c.tx.Exec(deleteQuery, id, cat.Id)
			if err != nil {
				return err
			}

		}

		for i, cat := range new {
			log.Println("Adding", cat.Name)
			err := c.associateBookAndCategoryType(idStr, cat, i)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if update.Authors != nil {
		old, err := c.GetCategoryTypesAssociatedWithBook(id.String(), Authors)
		if err != nil {
			return Book{}, false, err
		}
		err = handleCategories("DELETE FROM books_authors WHERE book_id = ? AND author_id = ?", *update.Authors, old, Authors)
		if err != nil {
			return Book{}, false, err
		}

		needsFileUpdate = true
	}

	if update.Genres != nil {
		old, err := c.GetCategoryTypesAssociatedWithBook(id.String(), Genres)
		if err != nil {
			return Book{}, false, err
		}
		err = handleCategories("DELETE FROM books_genres WHERE book_id = ? AND genre_id = ?", *update.Genres, old, Genres)
		if err != nil {
			return Book{}, false, err
		}
	}

	if update.Series != nil {
		old, err := c.GetCategoryTypesAssociatedWithBook(id.String(), Series)
		if err != nil {
			return Book{}, false, err
		}
		err = handleCategories("DELETE FROM books_series WHERE book_id = ? AND series_id = ?", *update.Series, old, Series)
		if err != nil {
			return Book{}, false, err
		}

		needsFileUpdate = true
	}

	if update.Narrators != nil {
		old, err := c.GetCategoryTypesAssociatedWithBook(id.String(), Narrators)
		if err != nil {
			return Book{}, false, err
		}
		err = handleCategories("DELETE FROM books_narrators WHERE book_id = ? AND narrator_id = ?", *update.Narrators, old, Narrators)
		if err != nil {
			return Book{}, false, err
		}
	}

	err := c.CleanupCategories()
	if err != nil {
		return Book{}, false, err
	}

	book, err := c.GetBook(id)
	if err != nil {
		return Book{}, false, err
	}

	needsFileUpdate = needsFileUpdate && book.Files.Root != nil

	if needsFileUpdate {
		authorDir, seriesDir, bookDir, err := c.GetPathComponents(id)
		if err != nil {
			return Book{}, false, err
		}

		book.Files.UpdateDirectory(path.Join(authorDir, seriesDir, bookDir))
		err = book.updateBookFiles(c.tx)
		if err != nil {
			return Book{}, false, err
		}
	}

	if indyTx {
		err = c.Commit()
		if err != nil {
			return Book{}, false, err
		}
	}

	log.Println("Updated book", id)

	return book, needsFileUpdate, nil
}

func (c Client) UpdateBookCover(id uuid.UUID, ext string) (string, string, error) {

	err := c.Begin()
	if err != nil {
		return "", "", err
	}
	defer c.Rollback()

	var dir *string
	var cover *string

	err = c.tx.QueryRow("SELECT directory, cover FROM books WHERE id = ?", id).Scan(&dir, &cover)
	if err != nil {
		return "", "", err
	}

	if dir == nil {
		return "", "", nil
	}
	if cover == nil {
		tmp := ""
		cover = &tmp
	}

	newCover := path.Join(*dir, "cover."+ext)
	_, err = c.tx.Exec("UPDATE books SET cover = ? WHERE id = ?", newCover, id)
	if err != nil {
		return "", "", err
	}

	err = c.Commit()
	if err != nil {
		return "", "", err
	}

	return *cover, newCover, nil
}

// Author -> Series -> Book Title
func (c Client) GetPathComponents(id uuid.UUID) (string, string, string, error) {

	indyTx := c.tx == nil
	if indyTx {
		err := c.Begin()
		if err != nil {
			return "", "", "", err
		}
		defer c.Rollback()
	}

	authorDir := "Unknown"
	authors, err := c.GetCategoryTypesAssociatedWithBook(id.String(), Authors)
	if err != nil {
		return "", "", "", err
	}
	if len(authors) > 0 {
		authorDir = authors[0].Name
	}

	seriesDir := ""
	indexStr := ""
	series, err := c.GetCategoryTypesAssociatedWithBook(id.String(), Series)
	if err != nil {
		return "", "", "", err
	}
	if len(series) > 0 {
		seriesDir = series[0].Name
		if series[0].Index != nil {
			indexStr = fmt.Sprintf("%s - ", *series[0].Index)
		}
	}

	title := ""
	err = c.tx.QueryRow("SELECT title FROM books WHERE id = ?", id).Scan(&title)
	if err != nil {
		return "", "", "", err
	}
	title = indexStr + title

	if indyTx {
		err := c.Commit()
		if err != nil {
			return "", "", "", err
		}
	}

	return authorDir, seriesDir, title, nil
}

func (c Client) DeleteBook(id uuid.UUID) error {

	err := c.Begin()
	if err != nil {
		return err
	}
	defer c.Rollback()

	_, err = c.tx.Exec("DELETE FROM books WHERE id = ?", id)
	if err != nil {
		return err
	}

	err = c.CleanupCategories()
	if err != nil {
		return err
	}

	err = c.Commit()
	if err != nil {
		return err
	}

	log.Println("Removed the book with the id \"", id, "\" from the database")

	return nil
}

func (c Client) GetBookDirectory(id uuid.UUID) (*string, error) {

	indyTx := c.tx == nil
	if indyTx {
		err := c.Begin()
		if err != nil {
			return nil, err
		}
		defer c.Rollback()
	}

	var dir *string
	err := c.tx.QueryRow("SELECT directory FROM books WHERE id = ?", id).Scan(&dir)
	if err != nil {
		return nil, err
	}

	if indyTx {
		err := c.Commit()
		if err != nil {
			return nil, err
		}
	}

	return dir, nil
}

func (c Client) DeleteBookFilesFromDatabase(id uuid.UUID) error {

	indyTx := c.tx == nil
	if indyTx {
		err := c.Begin()
		if err != nil {
			return err
		}
		defer c.Rollback()
	}

	_, err := c.tx.Exec("UPDATE books SET directory = NULL, audio_files = NULL, text_files = NULL, cover = NULL WHERE id = ? ", id)
	if err != nil {
		return err
	}

	if indyTx {
		err = c.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

// #region Book Methods

func (book *Book) getBookCategories(c Client) error {

	var err error

	book.Authors, err = c.GetCategoryTypesAssociatedWithBook(book.Id.String(), Authors)
	if err != nil {
		return err
	}

	book.Genres, err = c.GetCategoryTypesAssociatedWithBook(book.Id.String(), Genres)
	if err != nil {
		return err
	}

	book.Series, err = c.GetCategoryTypesAssociatedWithBook(book.Id.String(), Series)
	if err != nil {
		return err
	}

	book.Narrators, err = c.GetCategoryTypesAssociatedWithBook(book.Id.String(), Narrators)
	if err != nil {
		return err
	}

	return nil
}

func (book *Book) updateBookFiles(tx *sql.Tx) error {

	if tx == nil {
		return fmt.Errorf("updateBookFiles requires an active tx")
	}

	audio, text, err := book.Files.FileListsToJson()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
	UPDATE books 
	SET 
		updated_at = CURRENT_TIMESTAMP,
		directory = ?,
		audio_files = ?,
		text_files = ?,
		cover = ?
	WHERE id = ?`, book.Files.Root, audio, text, book.Files.Cover, book.Id)
	if err != nil {
		return err
	}

	return nil
}
