package metadata

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

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
	} `json:"docs"`
}

func SearchOpenLibrary(params SearchParams) (SearchResults, error) {

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
	if params.Sort != nil {
		q.Add("sort", *params.Sort)
	}

	u.RawQuery = q.Encode()

	log.Println("Querying OpenLibrary:", u.String())

	resp, err := http.Get(u.String())
	if err != nil {
		return SearchResults{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SearchResults{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var results OpenLibrarySearchResults
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return SearchResults{}, err
	}

	return results.Parse(params.Genres), nil
}

func (results *OpenLibrarySearchResults) Parse(genres *[]string) SearchResults {

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

		book := database.Book{
			Title:   result.Title,
			Year:    &result.FirstPublishYear,
			Authors: authors,
		}

		if len(genresCats) > 0 {
			book.Genres = genresCats
		}

		standardResults.Items = append(standardResults.Items, book)
	}

	return standardResults
}
