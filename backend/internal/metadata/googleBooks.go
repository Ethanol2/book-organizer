package metadata

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/Ethanol2/book-organizer/internal/cache"
	"github.com/Ethanol2/book-organizer/internal/database"
)

type GoogleBooksSearchResults struct {
	Kind       string            `json:"kind"`
	TotalItems int               `json:"totalItems"`
	Items      []GoogleBooksItem `json:"items"`
}

type GoogleBooksItem struct {
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
}

func SearchGoogleBooks(params SearchParams, key string, cache *cache.Cache) (SearchResults, error) {

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

	pageOffset := 0
	if params.Page != nil {

		limit := 10
		if params.Limit != nil {
			limit = *params.Limit
		}

		pageOffset = (*params.Page - 1) * limit
		log.Println(pageOffset)
		q.Add("startIndex", fmt.Sprint(pageOffset))
	}

	if params.Sort != nil {
		q.Add("orderBy", *params.Sort)
	}

	q.Add("key", key)

	u.RawQuery = q.Encode()

	log.Println("Querying GoogleBooks:", u.String())

	body, err := cache.HttpGet(u.String())
	if err != nil {
		return SearchResults{}, err
	}

	var results GoogleBooksSearchResults
	err = json.Unmarshal(body, &results)
	if err != nil {
		return SearchResults{}, err
	}

	return results.ParseSearch(pageOffset), nil
}

func (results *GoogleBooksSearchResults) ParseSearch(offset int) SearchResults {

	standardResults := SearchResults{
		TotalCount: results.TotalItems,
		Count:      len(results.Items),
		Offset:     offset,
	}

	for _, result := range results.Items {
		book, err := result.Parse()
		if err != nil {
			log.Println(err)
			continue
		}
		standardResults.Items = append(standardResults.Items, book)
	}

	return standardResults
}

func (result *GoogleBooksItem) Parse() (database.BookParams, error) {
	var year int
	var err error

	if dateStr := result.VolumeInfo.PublishedDate; dateStr != "" {
		if split := strings.Split(dateStr, "-"); len(split) > 1 {
			dateStr = split[0]
		}

		year, err = strconv.Atoi(dateStr)
		if err != nil {
			log.Println(err)
		}
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

	return database.BookParams{
		Title:       &result.VolumeInfo.Title,
		Subtitle:    &result.VolumeInfo.Subtitle,
		Description: &result.VolumeInfo.Description,
		Year:        &year,
		Publisher:   &result.VolumeInfo.Publisher,
		ISBN:        &isbn,
		MetadataKey: &result.ID,
		Authors:     &authors,
		Genres:      &genres,
		Cover:       &result.VolumeInfo.ImageLinks.Thumbnail,
	}, nil
}
