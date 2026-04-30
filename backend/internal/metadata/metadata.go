package metadata

import (
	"fmt"
	"regexp"
	"strconv"
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
	ASIN      *string   `json:"asin"`

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

// Matches AudioBookshelf's metadata format
type MetadataFile struct {
	Tags     []string `json:"tags"`
	Chapters []struct {
		ID    int     `json:"id"`
		Start int     `json:"start"`
		End   float64 `json:"end"`
		Title string  `json:"title"`
	} `json:"chapters,omitempty"`
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
	Explicit      bool     `json:"explicit,omitempty"`
	Abridged      bool     `json:"abridged,omitempty"`
}

func BookToMetadata(book database.Book) MetadataFile {
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

	series := []string{}
	for _, item := range book.Series {
		s := item.Name
		if item.Index != nil {
			s += fmt.Sprintf(" #%s", *item.Index)
		}
		series = append(series, s)
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

func MetadataToBook(metadata MetadataFile) database.Book {

	genres := database.StrToCategorySlice(metadata.Genres)
	authors := database.StrToCategorySlice(metadata.Authors)
	narrators := database.StrToCategorySlice(metadata.Narrators)

	series := []database.Category{}
	for _, s := range metadata.Series {
		split := strings.SplitN(s, "#", 1)
		cat := database.Category{
			Name: strings.TrimSpace(split[0]),
		}
		if len(split) > 1 {
			index := strings.TrimSpace(split[1])
			cat.Index = &index
		}
		series = append(series, cat)
	}

	var desc *string
	if len(metadata.Description) > 0 && metadata.Description != "null" {
		desc = &metadata.Description
	}

	var year *int
	if len(metadata.PublishedYear) > 0 && metadata.PublishedYear != "null" {
		y, err := strconv.Atoi(metadata.PublishedYear)
		if err == nil {
			year = &y
		} else {
			year = nil
		}
	}

	var isbn *string
	if len(metadata.Isbn) > 0 && metadata.Isbn != "null" {
		isbn = &metadata.Isbn
	}

	var asin *string
	if len(metadata.Isbn) > 0 && metadata.Asin != "null" {
		asin = &metadata.Asin
	}

	return database.Book{
		Title:       metadata.Title,
		Subtitle:    metadata.Subtitle,
		Description: desc,
		Publisher:   &metadata.Publisher,
		Year:        year,
		ISBN:        isbn,
		ASIN:        asin,

		Authors:   authors,
		Series:    series,
		Genres:    genres,
		Narrators: narrators,
		Tags:      metadata.Tags,
	}
}

func stripTags(s string) string {

	s = strings.ReplaceAll(s, "<br>", "\n")

	// Regular expression to match any content within brackets
	re := regexp.MustCompile(`<.*?>`)
	return re.ReplaceAllString(s, "")
}
