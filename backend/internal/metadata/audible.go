package metadata

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
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

type AudibleError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details struct {
			Asin string `json:"asin"`
			Code string `json:"code"`
		} `json:"details"`
	} `json:"error"`
}

var audibleRegions = map[string]string{"co.au": "au", "ca": "ca", "de": "de", "es": "es", "fr": "fr", "co.in": "in", "it": "it", "co.jp": "jp", "com": "us", "co.uk": "uk"}

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
	if params.Keywords != nil {
		for _, keyword := range *params.Keywords {
			q.Add("keywords", keyword)
		}
	}
	if params.Sort != nil {
		q.Add("products_sort_by", *params.Sort)
	}
	if params.Page != nil {
		q.Add("page", fmt.Sprint(*params.Page))
	}
	if params.Limit != nil {
		q.Add("num_results", fmt.Sprint(*params.Limit))
	}

	u.RawQuery = q.Encode()

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
			log.Println(err)
			continue
		}
		fullResults = append(fullResults, params)
	}

	return SearchResults{
		TotalCount: results.TotalResults,
		Count:      len(fullResults),
		Items:      fullResults,
	}, nil
}

func GetFromAudible(asin, region string, cache *cache.Cache) (database.BookParams, error) {

	if !IsValidASIN(asin) {
		return database.BookParams{}, fmt.Errorf("invalid asin")
	}

	u := url.URL{
		Scheme: "https",
		Host:   "api.audnex.us",
		Path:   fmt.Sprintf("books/%s", asin),
	}
	q := u.Query()
	q.Add("region", audibleRegions[region])
	u.RawQuery = q.Encode()

	log.Println("Querying Audible =>", u.String())

	key := fmt.Sprintf("api/metadata/%s?source=audible&region=%s", asin, region)

	body, err := cache.HttpGet(u.String())
	if err != nil {
		if len(body) > 0 {
			log.Println(err)
			var auErr AudibleError
			err := json.Unmarshal(body, &auErr)
			if err != nil {
				return database.BookParams{}, err
			}
			title := "Error retrieving details"
			return database.BookParams{
				Title:       &title,
				ASIN:        &asin,
				Description: &auErr.Error.Message,
				Key:         &key,
			}, nil
		}
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

	var isbn *string
	if IsValidISBN13(item.Isbn) {
		isbn = &item.Isbn
	}

	return database.BookParams{
		Title:       &item.Title,
		Description: &item.Description,
		Year:        &year,
		Publisher:   &item.PublisherName,
		ISBN:        isbn,
		ASIN:        &asin,
		Genres:      &genres,
		Series:      &series,
		Authors:     &authors,
		Narrators:   &narrators,
		Cover:       &item.Image,
		Key:         &key,
	}, nil
}

func IsValidAudibleRegion(region string) bool {
	for key := range audibleRegions {
		if region == key {
			return true
		}
	}
	return false
}

// Provided by Gemini
// IsValidASIN checks if a string is a valid Amazon Standard Identification Number.
func IsValidASIN(asin string) bool {
	asin = strings.ToUpper(strings.TrimSpace(asin))

	// ASIN must be exactly 10 alphanumeric characters
	match, _ := regexp.MatchString("^[A-Z0-9]{10}$", asin)
	if !match {
		return false
	}

	// If it starts with 'B', it's a standard ASIN (non-book)
	// These don't have a public checksum formula, so format is enough.
	if strings.HasPrefix(asin, "B") {
		return true
	}

	// If it doesn't start with 'B', it's likely an ISBN-10 (used for books)
	// We should validate the ISBN-10 checksum.
	return isValidISBN10(asin)
}

func isValidISBN10(isbn string) bool {
	if len(isbn) != 10 {
		return false
	}

	sum := 0
	for i := 0; i < 9; i++ {
		digit, err := strconv.Atoi(string(isbn[i]))
		if err != nil {
			return false // Must be digits 0-9
		}
		sum += digit * (10 - i)
	}

	// The last character can be 'X' (representing 10) or a digit
	lastChar := isbn[9]
	var lastValue int
	if lastChar == 'X' {
		lastValue = 10
	} else {
		var err error
		lastValue, err = strconv.Atoi(string(lastChar))
		if err != nil {
			return false
		}
	}

	sum += lastValue
	return sum%11 == 0
}
