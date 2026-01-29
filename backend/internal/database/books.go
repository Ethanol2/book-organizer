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

func (c Client) AddBook(params CreateBookParams) (Book, error) {

	id := uuid.New()

	tagsJson, err := json.Marshal(params.Tags)
	if err != nil {
		return Book{}, err
	}

	query := `
	INSERT INTO books
		(id, title, publish_year, description, tags, isbn, asin, publisher, audio_files, text_files, cover, created_at, updated_at)
	VALUES
		(?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, NULL, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)	
	`

	_, err = c.db.Exec(query, id, params.Title, params.Year, params.Description, tagsJson, params.ISBN, params.ASIN, params.Publisher)
	if err != nil {
		return Book{}, err
	}

	log.Println("Added \"", params.Title, "\" to books")

	sortCats := func(catType CategoryType, cats []Category) {
		for _, cat := range cats {

			if cat.Id == nil {
				cat, err = c.AddCategory(catType, cat.Value)
				if err != nil {
					log.Println(err)
					continue
				}
			}

			err := c.associateBookAndCategoryType(id.String(), cat)
			if err != nil {
				log.Println(err)
			}
		}
	}

	sortCats(Series, params.Series)
	sortCats(Genres, params.Genres)
	sortCats(Narrators, params.Narrators)
	sortCats(Authors, params.Authors)

	return Book{}, nil
}

func (c Client) GetBook(id uuid.UUID) (Book, error) {

	var book Book
	var tagsStr string
	var audioStr string
	var textStr string

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

	err = json.Unmarshal([]byte(tagsStr), &book.Tags)
	if err != nil {
		return Book{}, err
	}

	err = book.Files.ParseAudioJson(audioStr)
	if err != nil {
		return Book{}, err
	}

	err = book.Files.ParseTextJson(textStr)
	if err != nil {
		return Book{}, err
	}

	book.Authors, err = c.getCategoryTypesAssociatedWithBook(id.String(), Authors)
	if err != nil {
		return Book{}, err
	}

	book.Genres, err = c.getCategoryTypesAssociatedWithBook(id.String(), Genres)
	if err != nil {
		return Book{}, err
	}

	book.Series, err = c.getCategoryTypesAssociatedWithBook(id.String(), Series)
	if err != nil {
		return Book{}, err
	}

	book.Narrators, err = c.getCategoryTypesAssociatedWithBook(id.String(), Narrators)
	if err != nil {
		return Book{}, err
	}

	return book, nil
}
