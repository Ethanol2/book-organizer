package metadata

import (
	"fmt"

	"github.com/Ethanol2/book-organizer/internal/database"
)

type SearchParams struct {
	Title     *string   `json:"title"`
	Author    *string   `json:"author"`
	Year      *string   `json:"year"`
	Publisher *string   `json:"publisher"`
	Genres    *[]string `json:"genres"`
	Languages *[]string `json:"languages"`
	ISBN      *string   `json:"isbn"`

	Page *int    `json:"page"`
	Sort *string `json:"sort"`
}

type SearchResults struct {
	TotalCount int                   `json:"total_count"`
	Count      int                   `json:"count"`
	Offset     int                   `json:"offset"`
	Items      []database.BookParams `json:"items"`
}

// Matches AudioBookshelf's metadata format
type MetadataFile struct {
	Tags     []string `json:"tags"`
	Chapters []struct {
		ID    int     `json:"id"`
		Start int     `json:"start"`
		End   float64 `json:"end"`
		Title string  `json:"title"`
	} `json:"chapters"`
	Title         string   `json:"title"`
	Subtitle      *string  `json:"subtitle,omitempty"`
	Authors       []string `json:"authors"`
	Narrators     []string `json:"narrators"`
	Series        []string `json:"series"`
	Genres        []string `json:"genres"`
	PublishedYear string   `json:"publishedYear"`
	PublishedDate *string  `json:"publishedDate"`
	Publisher     string   `json:"publisher"`
	Description   string   `json:"description"`
	Isbn          string   `json:"isbn"`
	Asin          string   `json:"asin"`
	Language      string   `json:"language"`
	Explicit      bool     `json:"explicit"`
	Abridged      bool     `json:"abridged"`
}

func MetadataFileFromBook(book database.Book) MetadataFile {
	md := MetadataFile{}

	md.Title = book.Title
	md.Subtitle = book.Subtitle
	md.Authors = database.CategoryToStrSlice(book.Authors)
	md.Narrators = database.CategoryToStrSlice(book.Narrators)
	md.Series = database.CategoryToStrSlice(book.Series)
	md.Genres = database.CategoryToStrSlice(book.Genres)
	md.PublishedYear = fmt.Sprint(book.Year)
	md.Publisher = book.Publisher
	md.Description = book.Description
	md.Isbn = book.ISBN
	md.Asin = book.ASIN

	return md
}
