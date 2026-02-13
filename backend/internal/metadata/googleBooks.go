package metadata

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Ethanol2/book-organizer/internal/database"
)

type GoogleBooksSearchResults struct {
	Kind       string `json:"kind"`
	TotalItems int    `json:"totalItems"`
	Items      []struct {
		Kind       string `json:"kind"`
		ID         string `json:"id"`
		Etag       string `json:"etag"`
		SelfLink   string `json:"selfLink"`
		VolumeInfo struct {
			Title               string   `json:"title"`
			Subtitle            string   `json:"subtitle"`
			Authors             []string `json:"authors"`
			Publisher           string   `json:"publisher"`
			PublishedDate       string   `json:"publishedDate"`
			Description         string   `json:"description"`
			IndustryIdentifiers []struct {
				Type       string `json:"type"`
				Identifier string `json:"identifier"`
			} `json:"industryIdentifiers"`
			PageCount      int      `json:"pageCount"`
			PrintType      string   `json:"printType"`
			Categories     []string `json:"categories"`
			MaturityRating string   `json:"maturityRating"`
			ImageLinks     struct {
				SmallThumbnail string `json:"smallThumbnail"`
				Thumbnail      string `json:"thumbnail"`
			} `json:"imageLinks"`
			Language            string `json:"language"`
			PreviewLink         string `json:"previewLink"`
			InfoLink            string `json:"infoLink"`
			CanonicalVolumeLink string `json:"canonicalVolumeLink"`
		} `json:"volumeInfo"`
	} `json:"items"`
}

func SearchGoogleBooks(params SearchParams, key string) (SearchResults, error) {

	u := url.URL{
		Scheme: "https",
		Host:   "www.googleapis.com",
		Path:   "books/v1/volumes",
	}

	searchItems := []string{}
	if params.Title != nil {
		searchItems = append(searchItems, "intitle:"+*params.Title)
	}
	if params.Author != nil {
		searchItems = append(searchItems, "inauthor:"+*params.Author)
	}
	if params.Publisher != nil {
		searchItems = append(searchItems, "inpublisher:"+*params.Publisher)
	}
	if params.ISBN != nil {
		searchItems = append(searchItems, "isbn:"+*params.ISBN)
	}
	if params.Genres != nil {
		subjects := ""
		for _, subject := range *params.Genres {
			subjects += "subject:" + subject + " "
		}
		searchItems = append(searchItems, subjects[:len(subjects)-1])
	}

	q := u.Query()
	q.Add("q", strings.Join(searchItems, " "))

	pageOffset := (*params.Page - 1) * 10
	if params.Page != nil {
		q.Add("startIndex", fmt.Sprint(pageOffset))
	}
	if params.Sort != nil {
		q.Add("orderBy", *params.Sort)
	}

	q.Add("key", key)

	u.RawQuery = q.Encode()

	log.Println("Querying GoogleBooks:", u.String())

	resp, err := http.Get(u.String())
	if err != nil {
		return SearchResults{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return SearchResults{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var results GoogleBooksSearchResults
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return SearchResults{}, err
	}

	return results.Parse(pageOffset), nil
}

func (results *GoogleBooksSearchResults) Parse(offset int) SearchResults {

	standardResults := SearchResults{
		TotalCount: results.TotalItems,
		Count:      len(results.Items),
		Offset:     offset,
	}

	for _, result := range results.Items {

		year, err := strconv.Atoi(result.VolumeInfo.PublishedDate)
		if err != nil {
			log.Println(err)
			continue
		}

		isbn := ""
		for _, id := range result.VolumeInfo.IndustryIdentifiers {
			if id.Type == "ISBN_13" {
				isbn = id.Identifier
				break
			}
		}

		authors := []database.Category{}
		for _, author := range result.VolumeInfo.Authors {
			authors = append(authors, database.Category{
				Name: author,
			})
		}

		genres := []database.Category{}
		for _, genre := range result.VolumeInfo.Categories {
			genres = append(genres, database.Category{
				Name: genre,
			})
		}

		book := database.Book{
			Title:       result.VolumeInfo.Title,
			Subtitle:    &result.VolumeInfo.Subtitle,
			Description: result.VolumeInfo.Description,
			Year:        &year,
			Publisher:   result.VolumeInfo.Publisher,
			ISBN:        isbn,
			Authors:     authors,
			Genres:      genres,
		}

		standardResults.Items = append(standardResults.Items, book)
	}

	return standardResults
}
