package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Ethanol2/book-organizer/internal/database"
	"github.com/Ethanol2/book-organizer/internal/metadata"
)

func (cfg *apiConfig) handlerMetadataSearch(w http.ResponseWriter, r *http.Request) {

	// GET /books/search?source=google&title=harry+potter

	var searchParams metadata.SearchParams

	if title := r.URL.Query().Get("title"); title != "" {
		searchParams.Title = &title
	}
	if author := r.URL.Query().Get("author"); author != "" {
		searchParams.Author = &author
	}
	if year := r.URL.Query().Get("year"); year != "" {
		searchParams.Year = &year
	}
	if pub := r.URL.Query().Get("publisher"); pub != "" {
		searchParams.Publisher = &pub
	}
	if isbn := r.URL.Query().Get("isbn"); isbn != "" {
		searchParams.ISBN = &isbn
	}
	if page := r.URL.Query().Get("page"); page != "" {
		num, err := strconv.Atoi(page)
		if err == nil {
			searchParams.Page = &num
		}
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		num, err := strconv.Atoi(limit)
		if err == nil {
			searchParams.Limit = &num
		}
	}

	if genres := r.URL.Query()["genre"]; len(genres) > 0 {
		searchParams.Genres = &genres
	}
	if keywords := r.URL.Query()["keywords"]; len(keywords) > 0 {
		searchParams.Keywords = &keywords
	}
	if langs := r.URL.Query()["language"]; len(langs) > 0 {
		searchParams.Languages = &langs
	}

	var results metadata.SearchResults
	var err error

	switch r.URL.Query().Get("source") {

	case "":
		respondWithError(w, http.StatusBadRequest, MetadataSourceError, fmt.Errorf(MetadataSourceError))
		return

	case "open library":
		results, err = metadata.SearchOpenLibrary(searchParams, &cfg.mdCache)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, MetadataFetchError, err)
			return
		}

	case "google books":
		if cfg.googleBooksApiKey == "" {
			respondWithError(w, http.StatusInternalServerError, MetadataApiKeyMissing, fmt.Errorf(MetadataApiKeyMissing))
			return
		}

		results, err = metadata.SearchGoogleBooks(searchParams, cfg.googleBooksApiKey, &cfg.mdCache)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, MetadataFetchError, err)
			return
		}

	case "audible":
		region := r.URL.Query().Get("region")
		if region == "" || !metadata.IsValidAudibleRegion(region) {
			respondWithError(w, http.StatusBadRequest, "Querying audible requires a valid region. Valid regions are: co.au, ca, de, es, fr, co.in, it, co.jp, com, co.uk", fmt.Errorf("no valid region provided => %s", region))
			return
		}

		results, err = metadata.SearchAudible(searchParams, region, &cfg.mdCache)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, MetadataFetchError, err)
			return
		}

	default:
		src := r.URL.Query().Get("source")
		respondWithError(w, http.StatusBadRequest, MetadataSourceError, fmt.Errorf("Invalid source: %s", src))
		return
	}

	respondWithJson(w, http.StatusOK, results)
}

func (cfg *apiConfig) handlerGetMetadataBookDetails(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondWithError(w, http.StatusBadRequest, "Request missing id", fmt.Errorf("request missing id"))
		return
	}

	var result database.BookParams
	var err error
	switch r.URL.Query().Get("source") {

	case "":
		respondWithError(w, http.StatusBadRequest, MetadataSourceError, fmt.Errorf(MetadataSourceError))
		return

	case "open library":
		result, err = metadata.GetFromOpenLibrary(id, &cfg.mdCache)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, MetadataFetchError, err)
			return
		}

	case "google books":
		if cfg.googleBooksApiKey == "" {
			respondWithError(w, http.StatusInternalServerError, MetadataApiKeyMissing, fmt.Errorf(MetadataApiKeyMissing))
			return
		}

		result, err = metadata.GetFromGoogleBooks(id, cfg.googleBooksApiKey, &cfg.mdCache)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, MetadataFetchError, err)
			return
		}

	case "audible":
		region := r.URL.Query().Get("region")
		if region == "" || !metadata.IsValidAudibleRegion(region) {
			respondWithError(w, http.StatusBadRequest, "Querying audible requires a valid region. Valid regions are: au, ca, de, es, fr, in, it, jp, us, uk", fmt.Errorf("no valid region provided"))
			return
		}

		result, err = metadata.GetFromAudible(id, region, &cfg.mdCache)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, MetadataFetchError, err)
			return
		}

	default:
		src := r.URL.Query().Get("source")
		respondWithError(w, http.StatusBadRequest, MetadataSourceError, fmt.Errorf("invalid source: %s", src))
		return
	}

	respondWithJson(w, http.StatusOK, result)
}
