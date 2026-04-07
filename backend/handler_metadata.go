package main

import (
	"fmt"
	"net/http"

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

	if genres := r.URL.Query()["genre"]; len(genres) > 0 {
		searchParams.Genres = &genres
	}
	if langs := r.URL.Query()["language"]; len(langs) > 0 {
		searchParams.Languages = &langs
	}

	var results metadata.SearchResults
	var err error

	switch r.URL.Query().Get("source") {

	case "":
		respondWithError(w, http.StatusBadRequest, "Missing source", fmt.Errorf("request missing source in url"))
		return

	case "open library":
		results, err = metadata.SearchOpenLibrary(searchParams, &cfg.mdCache)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "something went wrong while querying openlibrary", err)
			return
		}

	case "google books":
		if cfg.googleBooksApiKey == "" {
			respondWithError(w, http.StatusInternalServerError, "missing google books api key in backend setup", fmt.Errorf("missing google books api key in backend setup"))
			return
		}

		results, err = metadata.SearchGoogleBooks(searchParams, cfg.googleBooksApiKey, &cfg.mdCache)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "something went wrong while querying google books", err)
			return
		}

	default:
		src := r.URL.Query().Get("source")
		respondWithError(w, http.StatusBadRequest, "Unknown source: "+src, fmt.Errorf("unknown source"))
		return
	}

	respondWithJson(w, http.StatusOK, results)
}

func (cfg *apiConfig) handlerGetBookDetails(w http.ResponseWriter, r *http.Request) {

}
