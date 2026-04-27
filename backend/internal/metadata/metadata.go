package metadata

import (
	"fmt"
	"regexp"
	"strings"

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

	Page  *int    `json:"page"`
	Limit *int    `json:"limit"`
	Sort  *string `json:"sort"`
}

type SearchResults struct {
	TotalCount int                   `json:"total_count"`
	Count      int                   `json:"count"`
	Offset     int                   `json:"offset"`
	Items      []database.BookParams `json:"items"`
}

type MetadataFileSeries struct {
	Series   string `json:"series"`
	Sequence string `json:"sequence,omitempty"`
}

// Matches AudioBookshelf's metadata format
type MetadataFile struct {
	Tags     []string `json:"tags"`
	Chapters []struct {
		ID    int     `json:"id"`
		Start int     `json:"start"`
		End   float64 `json:"end"`
		Title string  `json:"title"`
	} `json:"chapters,omitempty"`
	Title         string               `json:"title"`
	Subtitle      *string              `json:"subtitle,omitempty"`
	Authors       []string             `json:"authors"`
	Narrators     []string             `json:"narrators"`
	Series        []MetadataFileSeries `json:"series"`
	Genres        []string             `json:"genres"`
	PublishedYear string               `json:"publishedYear"`
	PublishedDate *string              `json:"publishedDate"`
	Publisher     string               `json:"publisher"`
	Description   string               `json:"description"`
	Isbn          string               `json:"isbn"`
	Asin          string               `json:"asin"`
	Language      string               `json:"language"`
	Explicit      bool                 `json:"explicit,omitempty"`
	Abridged      bool                 `json:"abridged,omitempty"`
}

func MetadataFileFromBook(book database.Book) MetadataFile {
	md := MetadataFile{}

	pub := ""
	if book.Publisher != nil {
		pub = *book.Publisher
	}

	desc := ""
	if book.Description != nil {
		desc = *book.Description
	}

	isbn := ""
	if book.ISBN != nil {
		isbn = *book.ISBN
	}

	asin := ""
	if book.ASIN != nil {
		asin = *book.ASIN
	}

	year := ""
	if book.Year != nil {
		year = fmt.Sprint(*book.Year)
	}

	series := []MetadataFileSeries{}
	for _, item := range book.Series {
		sequence := ""
		if item.Index != nil {
			sequence = fmt.Sprint(*item.Index)
		}
		series = append(series, MetadataFileSeries{Series: item.Name, Sequence: sequence})
	}

	md.Title = book.Title
	md.Subtitle = book.Subtitle
	md.Authors = database.CategoryToStrSlice(book.Authors)
	md.Narrators = database.CategoryToStrSlice(book.Narrators)
	md.Series = series
	md.Genres = database.CategoryToStrSlice(book.Genres)
	md.PublishedYear = year
	md.Publisher = pub
	md.Description = desc
	md.Isbn = isbn
	md.Asin = asin
	md.Tags = book.Tags

	return md
}

func stripTags(s string) string {

	s = strings.ReplaceAll(s, "<br>", "\n")

	// Regular expression to match any content within brackets
	re := regexp.MustCompile(`<.*?>`)
	return re.ReplaceAllString(s, "")
}
