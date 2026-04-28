package metadata

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/Ethanol2/book-organizer/internal/cache"
	"github.com/Ethanol2/book-organizer/internal/database"
)

type AudibleSearchResults struct {
	ProductFilters []interface{} `json:"product_filters"`
	Products       []struct {
		Asin string `json:"asin"`
	} `json:"products"`
	ResponseGroups []string `json:"response_groups"`
	TotalResults   int      `json:"total_results"`
}

type AudibleBookDetails struct {
	Asin             string    `json:"asin"`
	Isbn             string    `json:"isbn"`
	Copyright        int       `json:"copyright"`
	Description      string    `json:"description"`
	FormatType       string    `json:"formatType"`
	Image            string    `json:"image"`
	IsAdult          bool      `json:"isAdult"`
	Language         string    `json:"language"`
	LiteratureType   string    `json:"literatureType"`
	PublisherName    string    `json:"publisherName"`
	Rating           string    `json:"rating"`
	Region           string    `json:"region"`
	ReleaseDate      time.Time `json:"releaseDate"`
	RuntimeLengthMin int       `json:"runtimeLengthMin"`
	Summary          string    `json:"summary"`
	Title            string    `json:"title"`

	Authors []struct {
		Asin string `json:"asin"`
		Name string `json:"name"`
	} `json:"authors"`
	Genres []struct {
		Asin string `json:"asin"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"genres"`
	Narrators []struct {
		Name string `json:"name"`
	} `json:"narrators"`
	SeriesPrimary *struct {
		Asin     string `json:"asin"`
		Name     string `json:"name"`
		Position string `json:"position"`
	} `json:"seriesPrimary,omitempty"`
	SeriesSecondary *struct {
		Asin     string `json:"asin"`
		Name     string `json:"name"`
		Position string `json:"position"`
	} `json:"seriesSecondary,omitempty"`
}

var audibleRegions = [...]string{"au", "ca", "de", "es", "fr", "in", "it", "jp", "us", "uk"}

func SearchAudible(params SearchParams, region string, cache *cache.Cache) (SearchResults, error) {

	if params.ASIN != nil {
		book, err := GetFromAudible(*params.ASIN, region, cache)
		if err != nil {
			return SearchResults{}, err
		}
		return SearchResults{
			TotalCount: 1,
			Count:      1,
			Offset:     0,
			Items:      []database.BookParams{book},
		}, nil
	}

	u := url.URL{
		Scheme: "https",
		Host:   "api.audible." + region,
		Path:   "1.0/catalog/products",
	}
	q := u.Query()

	// author

	// keywords

	// num_results (int) – (max: 50)

	// page (int)

	// publisher

	// title

	if params.Title != nil {
		q.Add("title", *params.Title)
	}
	if params.Publisher != nil {
		q.Add("publisher", *params.Publisher)
	}
	if params.Author != nil {
		q.Add("author", *params.Author)
	}
	if params.Genres != nil {
		for _, genre := range *params.Genres {
			q.Add("keywords", genre)
		}
	}
	if params.Page != nil {
		q.Add("page", fmt.Sprint(*params.Page))
	}
	if params.Limit != nil {
		q.Add("num_results", fmt.Sprint(*params.Limit))
	}

	log.Println("Querying Audible =>", u.String())

	body, err := cache.HttpGet(u.String())
	if err != nil {
		return SearchResults{}, err
	}

	var results AudibleSearchResults
	err = json.Unmarshal(body, &results)
	if err != nil {
		return SearchResults{}, err
	}

	if results.TotalResults == 0 {
		return SearchResults{
			Count:      0,
			TotalCount: 0,
			Offset:     0,
			Items:      []database.BookParams{},
		}, nil
	}

	fullResults := []database.BookParams{}
	for _, result := range results.Products {
		params, err := GetFromAudible(result.Asin, region, cache)
		if err != nil {
			return SearchResults{}, err
		}
		fullResults = append(fullResults, params)
	}

	return SearchResults{
		TotalCount: results.TotalResults,
		Count:      len(fullResults),
	}, nil
}

func GetFromAudible(asin, region string, cache *cache.Cache) (database.BookParams, error) {

	u := url.URL{
		Scheme: "https",
		Host:   "api.audnex.us",
		Path:   fmt.Sprintf("books/%s", asin),
	}
	u.Query().Add("region", region)

	log.Println("Querying Audible =>", u.String())

	body, err := cache.HttpGet(u.String())
	if err != nil {
		return database.BookParams{}, err
	}

	var item AudibleBookDetails

	err = json.Unmarshal(body, &item)
	if err != nil {
		return database.BookParams{}, err
	}

	year := item.ReleaseDate.Year()

	genres := []database.Category{}
	for _, genre := range item.Genres {
		genres = append(genres, database.Category{
			Name: genre.Name,
		})
	}

	series := []database.Category{}
	if item.SeriesPrimary != nil {
		series = append(series, database.Category{
			Name:  item.SeriesPrimary.Name,
			Index: &item.SeriesPrimary.Position,
		})
	}
	if item.SeriesSecondary != nil {
		series = append(series, database.Category{
			Name:  item.SeriesPrimary.Name,
			Index: &item.SeriesPrimary.Position,
		})
	}

	authors := []database.Category{}
	for _, author := range item.Authors {
		authors = append(authors, database.Category{
			Name: author.Name,
		})
	}

	narrators := []database.Category{}
	for _, narrator := range item.Narrators {
		narrators = append(narrators, database.Category{
			Name: narrator.Name,
		})
	}

	key := fmt.Sprintf("api/metadata/%s?source=audible&region=%s", item.Asin, region)

	return database.BookParams{
		Title:       &item.Title,
		Description: &item.Description,
		Year:        &year,
		ISBN:        &item.Isbn,
		ASIN:        &item.Asin,
		Genres:      &genres,
		Series:      &series,
		Authors:     &authors,
		Narrators:   &narrators,
		Cover:       &item.Image,
		Key:         &key,
	}, nil
}

func IsValidAudibleRegion(region string) bool {
	for i := range audibleRegions {
		if region == audibleRegions[i] {
			return true
		}
	}
	return false
}
