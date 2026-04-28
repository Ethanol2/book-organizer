package metadata

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/Ethanol2/book-organizer/internal/cache"
	"github.com/Ethanol2/book-organizer/internal/database"
)

type OpenLibrarySearchResults struct {
	Start            int         `json:"start"`
	NumFoundExact    bool        `json:"numFoundExact"`
	NumFound         int         `json:"num_found"`
	DocumentationURL string      `json:"documentation_url"`
	Q                string      `json:"q"`
	Offset           interface{} `json:"offset"`
	Docs             []struct {
		AuthorKey          []string `json:"author_key,omitempty"`
		AuthorName         []string `json:"author_name,omitempty"`
		CoverEditionKey    string   `json:"cover_edition_key,omitempty"`
		CoverI             int      `json:"cover_i,omitempty"`
		EbookAccess        string   `json:"ebook_access"`
		EditionCount       int      `json:"edition_count"`
		FirstPublishYear   int      `json:"first_publish_year,omitempty"`
		HasFulltext        bool     `json:"has_fulltext"`
		Ia                 []string `json:"ia,omitempty"`
		IaCollection       []string `json:"ia_collection,omitempty"`
		Key                string   `json:"key"`
		Language           []string `json:"language,omitempty"`
		LendingEditionS    string   `json:"lending_edition_s,omitempty"`
		LendingIdentifierS string   `json:"lending_identifier_s,omitempty"`
		PublicScanB        bool     `json:"public_scan_b"`
		Title              string   `json:"title"`
		Subtitle           string   `json:"subtitle,omitempty"`
		SeriesName         []string `json:"series_name,omitempty"`
		SeriesPosition     []string `json:"series_position,omitempty"`
	} `json:"docs"`
}

type Description string
type OpenLibraryItem struct {
	Description Description `json:"description"`
	Title       string      `json:"title"`
	Subtitle    string      `json:"subtitle"`
	Subjects    []string    `json:"subjects"`
}

func SearchOpenLibrary(params SearchParams, cache *cache.Cache) (SearchResults, error) {

	u := url.URL{
		Scheme: "https",
		Host:   "openlibrary.org",
		Path:   "search.json",
	}

	searchItems := []string{}
	if params.Title != nil {
		searchItems = append(searchItems, "title:"+*params.Title)
	}
	if params.Author != nil {
		searchItems = append(searchItems, "author:"+*params.Author)
	}
	if params.Publisher != nil {
		searchItems = append(searchItems, "publisher:"+*params.Publisher)
	}
	if params.Year != nil {
		searchItems = append(searchItems, "publish_year:"+*params.Year)
	}
	if params.Genres != nil {
		subjects := "subject:"
		for _, subject := range *params.Genres {
			subjects += subject + " "
		}
		searchItems = append(searchItems, subjects[:len(subjects)-1])
	}
	if params.Languages != nil {
		languages := "language:"
		for _, lang := range *params.Languages {
			languages += lang + " "
		}
		searchItems = append(searchItems, languages[:len(languages)-1])
	}

	q := u.Query()
	q.Add("q", strings.Join(searchItems, " "))

	if params.Page != nil {
		q.Add("page", fmt.Sprint(*params.Page))
	}
	if params.Limit != nil {
		q.Add("limit", fmt.Sprint(*params.Limit))
	}
	if params.Sort != nil {
		q.Add("sort", *params.Sort)
	}

	u.RawQuery = q.Encode()

	log.Println("Querying OpenLibrary:", u.String())

	body, err := cache.HttpGet(u.String())
	if err != nil {
		return SearchResults{}, err
	}

	var results OpenLibrarySearchResults
	err = json.Unmarshal(body, &results)
	if err != nil {
		return SearchResults{}, err
	}

	return results.parse(params.Genres), nil
}

func GetFromOpenLibrary(id string, cache *cache.Cache) (database.BookParams, error) {
	u := url.URL{
		Scheme: "https",
		Host:   "openlibrary.org",
		Path:   fmt.Sprintf("works/%s.json", id),
	}

	log.Println("Querying OpenLibrary:", u.String())

	body, err := cache.HttpGet(u.String())
	if err != nil {
		return database.BookParams{}, err
	}

	var olItem OpenLibraryItem

	err = json.Unmarshal(body, &olItem)
	if err != nil {
		return database.BookParams{}, err
	}

	desc := stripTags(string(olItem.Description))

	genres := []database.Category{}
	series := []database.Category{}

	for _, genre := range olItem.Subjects {

		if len(genre) == 0 {
			continue
		}

		split := strings.Split(genre, ":")

		if len(split) == 1 {

			genres = append(genres, database.Category{
				Type: database.Genres,
				Name: strings.ReplaceAll(genre, ", ", "-"),
			})

		} else {
			switch split[0] {
			case "series":
			case "franchise":
				series = append(series, database.Category{
					Type: database.Series,
					Name: split[1],
				})
			case "genre":
				genres = append(genres, database.Category{
					Type: database.Genres,
					Name: strings.ReplaceAll(split[1], ", ", "-"),
				})
			}
		}
	}

	return database.BookParams{
		Description: &desc,
		Title:       &olItem.Title,
		Subtitle:    &olItem.Subtitle,
		Genres:      &genres,
	}, nil
}

func (results *OpenLibrarySearchResults) parse(genres *[]string) SearchResults {

	standardResults := SearchResults{
		TotalCount: results.NumFound,
		Count:      len(results.Docs),
		Offset:     results.Start,
	}

	genresCats := []database.Category{}
	if genres != nil {
		for _, genre := range *genres {
			genresCats = append(genresCats, database.Category{
				Name: genre,
			})
		}
	}

	for _, result := range results.Docs {

		authors := []database.Category{}
		for _, author := range result.AuthorName {
			authors = append(authors, database.Category{
				Name: author,
			})
		}

		seriesList := []database.Category{}
		for i := range result.SeriesName {
			series := database.Category{
				Name:  result.SeriesName[i],
				Index: &result.SeriesPosition[i],
			}
			seriesList = append(seriesList, series)
		}

		if result.LendingIdentifierS != "" {
			result.LendingIdentifierS = strings.Trim(result.LendingIdentifierS, "isbn_")
		}

		cover := fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-L.jpg", result.CoverI)
		key := fmt.Sprintf("/api/metadata/%s?source=%s", strings.Trim(result.Key, "/works/"), "open%20library")

		book := database.BookParams{
			Title:    &result.Title,
			Subtitle: &result.Subtitle,
			Year:     &result.FirstPublishYear,
			ISBN:     &result.LendingIdentifierS,
			Authors:  &authors,
			Series:   &seriesList,
			Cover:    &cover,
			Key:      &key,
		}

		if len(genresCats) > 0 {
			book.Genres = &genresCats
		}

		standardResults.Items = append(standardResults.Items, book)
	}

	return standardResults
}

func (d *Description) UnmarshalJSON(b []byte) error {
	// Try unmarshaling as a simple string first
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		*d = Description(s)
		return nil
	}

	// If that fails, try unmarshaling as the object struct
	var obj struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(b, &obj); err == nil {
		*d = Description(obj.Value)
		return nil
	}

	// Fallback for empty or null
	*d = ""
	return nil
}
