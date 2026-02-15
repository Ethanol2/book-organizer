package database

import (
	"database/sql"
	"encoding/json"
	"errors"
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
	Description string     `json:"description"`
	Year        *int       `json:"year"`
	ISBN        string     `json:"isbn"`
	ASIN        string     `json:"asin"`
	Tags        []string   `json:"tags"`
	Publisher   string     `json:"publisher"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`

	// Categories
	Series    []Category `json:"series"`
	Authors   []Category `json:"authors"`
	Genres    []Category `json:"genres"`
	Narrators []Category `json:"narrators"`

	Files BookFiles `json:"files,omitempty"`
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

	Cover *string `json:"cover"`
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

	c.Begin()
	defer c.Rollback()

	id := uuid.New()

	tagsJson, err := json.Marshal(params.Tags)
	if err != nil {
		return Book{}, err
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

	log.Println("Added \"", params.Title, "\" to books")

	sortCats := func(catType CategoryType, cats []Category) error {
		log.Println("Associating", catType)

		for i, cat := range cats {

			if cat.Id == nil {
				result, err := c.GetCategoryByValue(catType, cat.Name)

				if err != nil {
					return err
				}

				if result == (Category{}) {
					result, err = c.AddCategory(catType, cat.Name)
					if err != nil {
						return err
					}
				}
				result.Index = cat.Index
				cat = result
			}

			err := c.associateBookAndCategoryType(id.String(), cat, i)
			if err != nil {
				return err
			}
		}
		return nil
	}

	err = sortCats(Series, *params.Series)
	if err != nil {
		return Book{}, err
	}

	err = sortCats(Genres, *params.Genres)
	if err != nil {
		return Book{}, err
	}

	err = sortCats(Narrators, *params.Narrators)
	if err != nil {
		return Book{}, err
	}

	err = sortCats(Authors, *params.Authors)

	if err != nil {
		return Book{}, err
	}

	err = c.Commit()
	if err != nil {
		return Book{}, err
	}

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

	defer log.Println("Retrieved \"", book.Title, "\" from books")

	return book, nil
}

func (c Client) GetBooks() ([]Book, error) {

	c.Begin()
	defer c.Rollback()

	books := []Book{}

	rows, err := c.tx.Query("SELECT * FROM books")
	if err != nil {
		return []Book{}, err
	}

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
		)
		if err != nil {
			log.Println(err)
			continue
		}

		if audioStr != nil {
			err = book.Files.ParseAudioJson(*audioStr)
			if err != nil {
				return []Book{}, err
			}
		}

		if textStr != nil {
			err = book.Files.ParseTextJson(*textStr)
			if err != nil {
				return []Book{}, err
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
		return []Book{}, err
	}

	return books, nil
}

func (c Client) AssociateBookAndDownload(bookId, downloadId uuid.UUID, author, series string) (Book, error) {

	tx, err := c.db.Begin()
	if err != nil {
		return Book{}, err
	}
	defer tx.Rollback()

	var files BookFiles
	var Audio *string
	var Text *string

	err = tx.QueryRow(`
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

	files.Prepend(path.Join(author, series))
	audio, text, err := files.FileListsToJson()
	if err != nil {
		return Book{}, err
	}

	_, err = tx.Exec(`
	UPDATE books 
	SET 
		updated_at = CURRENT_TIMESTAMP,
		directory = ?,
		audio_files = ?,
		text_files = ?,
		cover = ?
	WHERE id = ?`, files.Root, audio, text, files.Cover, bookId)
	if err != nil {
		return Book{}, err
	}

	err = tx.Commit()
	if err != nil {
		return Book{}, err
	}

	return c.GetBook(bookId)
}

func (c Client) UpdateBook(id uuid.UUID, update BookParams) (Book, error) {

	err := c.Begin()
	if err != nil {
		return Book{}, err
	}
	defer c.Rollback()

	setParts := []string{"updated_at = CURRENT_TIMESTAMP"}
	args := []interface{}{}

	add := func(part string, arg interface{}) {
		setParts = append(setParts, part+" = ?")
		args = append(args, arg)
	}

	if update.Title != nil {
		add("title", update.Title)
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
			return Book{}, err
		}
		add("tags", tagsJson)
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
		query := "UPDATE books SET " + strings.Join(setParts, ", ") + "WHERE id = ?"
		args = append(args, id)
		_, err := c.tx.Exec(query, args...)
		if err != nil {
			return Book{}, err
		}
	}

	idStr := id.String()
	handleCategories := func(deleteQuery string, update, old []Category, catType CategoryType) error {
		new := []Category{}
		removed := old

		for _, cat := range update {
			if slices.Contains(old, cat) {
				index := slices.Index(removed, cat)
				removed = slices.Delete(removed, index, index+1)
			} else {
				cat.Type = catType
				new = append(new, cat)
			}
		}

		for _, cat := range removed {
			_, err := c.tx.Exec(deleteQuery, id, cat.Id)
			if err != nil {
				return err
			}
		}

		for i, cat := range new {
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
			return Book{}, err
		}
		err = handleCategories("DELETE FROM books_authors WHERE book_id = ? AND author_id = ?", *update.Authors, old, Authors)
		if err != nil {
			return Book{}, err
		}
	}

	if update.Genres != nil {
		old, err := c.GetCategoryTypesAssociatedWithBook(id.String(), Genres)
		if err != nil {
			return Book{}, err
		}
		err = handleCategories("DELETE FROM books_genres WHERE book_id = ? AND genre_id = ?", *update.Genres, old, Genres)
		if err != nil {
			return Book{}, err
		}
	}

	if update.Series != nil {
		old, err := c.GetCategoryTypesAssociatedWithBook(id.String(), Series)
		if err != nil {
			return Book{}, err
		}
		err = handleCategories("DELETE FROM books_series WHERE book_id = ? AND series_id = ?", *update.Series, old, Series)
		if err != nil {
			return Book{}, err
		}
	}

	if update.Narrators != nil {
		old, err := c.GetCategoryTypesAssociatedWithBook(id.String(), Narrators)
		if err != nil {
			return Book{}, err
		}
		err = handleCategories("DELETE FROM books_narrators WHERE book_id = ? AND narrator_id = ?", *update.Narrators, old, Narrators)
		if err != nil {
			return Book{}, err
		}
	}

	err = c.Commit()
	if err != nil {
		return Book{}, err
	}

	log.Println("Updated book \"", *update.Title, "\" (", id, ")")

	return c.GetBook(id)
}

func (c Client) UpdateBookCover(id uuid.UUID, ext string) (string, string, error) {

	var dir *string
	var cover *string

	err := c.tx.QueryRow("SELECT directory, cover FROM books WHERE id = ?", id).Scan(&dir, &cover)
	if err != nil {
		return "", "", err
	}

	if dir == nil {
		return "", "", nil
	}

	newCover := path.Join(*dir, "cover."+ext)
	_, err = c.tx.Exec("UPDATE books SET cover = ? WHERE id = ?", newCover, id)
	if err != nil {
		return "", "", err
	}

	return *cover, newCover, nil
}

func (c Client) GetPrimaryAuthorAndSeries(id uuid.UUID) (string, string, error) {
	authorDir := "Unknown"
	authors, err := c.GetCategoryTypesAssociatedWithBook(id.String(), Authors)
	if err != nil {
		return "", "", err
	}
	if len(authors) > 0 {
		authorDir = authors[0].Name
	}

	seriesDir := ""
	series, err := c.GetCategoryTypesAssociatedWithBook(id.String(), Series)
	if err != nil {
		return "", "", err
	}
	if len(series) > 0 {
		seriesDir = series[0].Name
	}

	return authorDir, seriesDir, nil
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
