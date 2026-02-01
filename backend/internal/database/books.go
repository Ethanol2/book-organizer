package database

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
)

type Book struct {
	Id          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Year        *int      `json:"year"`
	ISBN        string    `json:"isbn"`
	ASIN        string    `json:"asin"`
	Tags        []string  `json:"tags"`
	Publisher   string    `json:"publisher"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Categories
	Series    []Category `json:"series"`
	Authors   []Category `json:"authors"`
	Genres    []Category `json:"genres"`
	Narrators []Category `json:"narrators"`

	Files BookFiles `json:"files"`
}

type CreateBookParams struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Year        *int     `json:"year"`
	ISBN        string   `json:"isbn"`
	ASIN        string   `json:"asin"`
	Tags        []string `json:"tags"`
	Publisher   string   `json:"publisher"`

	// Categories
	Series    []Category `json:"series"`
	Authors   []Category `json:"authors"`
	Genres    []Category `json:"genres"`
	Narrators []Category `json:"narrators"`
}

func (c Client) CheckBookExists(id uuid.UUID) (bool, error) {
	var exists bool
	err := c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM books WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (c Client) AddBook(params CreateBookParams) (Book, error) {

	tx, _ := c.db.Begin()
	defer tx.Rollback()

	id := uuid.New()

	tagsJson, err := json.Marshal(params.Tags)
	if err != nil {
		return Book{}, err
	}

	query := `
	INSERT INTO books
		(id, title, publish_year, description, tags, isbn, asin, publisher, audio_files, text_files, cover, created_at, updated_at)
	VALUES
		(?, ?, ?, ?, ?, ?, ?, ?, NULL, NULL, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)	
	`

	_, err = tx.Exec(query, id, params.Title, params.Year, params.Description, tagsJson, params.ISBN, params.ASIN, params.Publisher)
	if err != nil {
		return Book{}, err
	}

	log.Println("Added \"", params.Title, "\" to books")

	sortCats := func(catType CategoryType, cats []Category) {
		log.Println("Associating", catType)

		for _, cat := range cats {
			if cat.Id == nil {
				value := cat.Name
				index := cat.Index
				cat, err = c.GetCategoryByValue(catType, value)

				if err != nil {
					log.Println(err)
					continue
				}

				if err == nil && cat == (Category{}) {
					cat, err = c.AddCategory(tx, catType, value)
					if err != nil {
						log.Println(err)
						continue
					}
				}
				cat.Index = index
			}

			err := c.associateBookAndCategoryType(tx, id.String(), cat)
			if err != nil {
				log.Println(err)
			}
		}
	}

	sortCats(Series, params.Series)
	sortCats(Genres, params.Genres)
	sortCats(Narrators, params.Narrators)
	sortCats(Authors, params.Authors)

	err = tx.Commit()
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
		&book.Year,
		&book.Description,
		&tagsStr,
		&book.ISBN,
		&book.ASIN,
		&book.Publisher,
		&audioStr,
		&textStr,
		&book.Files.Cover,
		&book.CreatedAt,
		&book.UpdatedAt,
	)
	if err != nil {
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

	books := []Book{}

	rows, err := c.db.Query("SELECT * FROM books")
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
			&book.Year,
			&book.Description,
			&tagsStr,
			&book.ISBN,
			&book.ASIN,
			&book.Publisher,
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

		if tagsStr != nil {
			err = json.Unmarshal([]byte(*tagsStr), &book.Tags)
			if err != nil {
				log.Println(err)
				continue
			}
		}

		if audioStr != nil {
			err = book.Files.ParseAudioJson(*audioStr)
			if err != nil {
				log.Println(err)
				continue
			}
		}

		if textStr != nil {
			err = book.Files.ParseTextJson(*textStr)
			if err != nil {
				log.Println(err)
				continue
			}
		}

		err = book.getBookCategories(c)
		if err != nil {
			log.Println(err)
			continue
		}

		books = append(books, book)
	}

	return books, nil
}

func (c Client) AssociateBookAndDownload(bookId, downloadId uuid.UUID) (Book, error) {

	tx, _ := c.db.Begin()
	defer tx.Rollback()

	_, err := tx.Exec(`
	UPDATE books
	SET
		audio_files = downloads.audio_files,
		text_files = downloads.text_files,
		cover = downloads.cover
	FROM downloads
	WHERE books.id = ? AND downloads.id = ?
	`, bookId, downloadId)
	if err != nil {
		return Book{}, err
	}

	_, err = tx.Exec("UPDATE books SET updated_at = CURRENT_TIMESTAMP WHERE id = ?", bookId)
	if err != nil {
		return Book{}, err
	}

	err = tx.Commit()
	if err != nil {
		return Book{}, err
	}

	return c.GetBook(bookId)
}

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
