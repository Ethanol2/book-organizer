package metadata

import "github.com/Ethanol2/book-organizer/internal/database"

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
	TotalCount int             `json:"total_count"`
	Count      int             `json:"count"`
	Offset     int             `json:"offset"`
	Items      []database.Book `json:"items"`
}
