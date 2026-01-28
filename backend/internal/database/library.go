package database

import "github.com/google/uuid"

type Book struct {
	Id          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Year        *int      `json:"year"`
	ISBN        string    `json:"isbn"`
	ASIN        string    `json:"asin"`
	Authors     []string  `json:"authors"`
	Series      []string  `json:"series"`
	SeriesIndex []int     `json:"series_index"`
	Genres      []string  `json:"genres"`
	Tags        []string  `json:"tags"`
	Narrators   []string  `json:"narrators"`
	Files       BookFiles `json:"files"`
}
