package metadata

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Ethanol2/book-organizer/internal/database"
	"github.com/Ethanol2/book-organizer/internal/fileManagement"
)

type SearchParams struct {
	Title     *string   `json:"title"`
	Author    *string   `json:"author"`
	Year      *string   `json:"year"`
	Publisher *string   `json:"publisher"`
	Genres    *[]string `json:"genres"`
	Languages *[]string `json:"languages"`
	Keywords  *[]string `json:"keywords"`
	ISBN      *string   `json:"isbn"`
	ASIN      *string   `json:"asin"`

	Page  *int    `json:"page"`
	Limit *int    `json:"limit"`
	Sort  *string `json:"sort"`
	Order *string `json:"order"`
}

type SearchResults struct {
	TotalCount int                   `json:"total_count"`
	Count      int                   `json:"count"`
	Offset     int                   `json:"offset"`
	Items      []database.BookParams `json:"items"`

	Error *string `json:"error,omitempty"`
}

func BookToMetadata(book database.Book) *fileManagement.MetadataFile {
	md := fileManagement.MetadataFile{}

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

	return &md
}

func MetadataToBookParams(metadata fileManagement.MetadataFile) database.BookParams {

	genres := database.StrToCategorySlice(metadata.Genres)
	authors := database.StrToCategorySlice(metadata.Authors)
	narrators := database.StrToCategorySlice(metadata.Narrators)

	series := []database.Category{}
	for _, s := range metadata.Series {
		split := strings.SplitN(s, "#", 2)
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

	return database.BookParams{
		Title:       &metadata.Title,
		Subtitle:    metadata.Subtitle,
		Description: desc,
		Publisher:   &metadata.Publisher,
		Year:        year,
		ISBN:        isbn,
		ASIN:        asin,

		Authors:   &authors,
		Series:    &series,
		Genres:    &genres,
		Narrators: &narrators,
		Tags:      &metadata.Tags,
	}
}

// Function provided by Gemini
// IsValidISBN13 validates the checksum of a 13-digit ISBN string.
func IsValidISBN13(isbn string) bool {
	// 1. Remove common formatting (hyphens and spaces)
	isbn = strings.ReplaceAll(isbn, "-", "")
	isbn = strings.ReplaceAll(isbn, " ", "")

	// 2. Validate format: exactly 13 digits
	match, _ := regexp.MatchString(`^\d{13}$`, isbn)
	if !match {
		return false
	}

	// 3. Calculate weighted sum
	sum := 0
	for i, char := range isbn {
		digit := int(char - '0')
		if i%2 == 0 {
			// Even index (1st, 3rd, 5th digit...) multiplier is 1
			sum += digit
		} else {
			// Odd index (2nd, 4th, 6th digit...) multiplier is 3
			sum += digit * 3
		}
	}

	// 4. Check if the final sum is a multiple of 10
	return sum%10 == 0
}

func stripTags(s string) string {

	s = strings.ReplaceAll(s, "<br>", "\n")

	// Regular expression to match any content within brackets
	re := regexp.MustCompile(`<.*?>`)
	return re.ReplaceAllString(s, "")
}
