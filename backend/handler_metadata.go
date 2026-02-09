package main

import (
	"fmt"
	"net/http"

	"github.com/Ethanol2/book-organizer/internal/metadata"
)

func (cfg *apiConfig) handlerSearchOpenLibrary(params metadata.SearchParams, w http.ResponseWriter, r *http.Request) {

	results, err := metadata.SearchOpenLibrary(params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "something went wrong while querying openlibrary", err)
		return
	}

	respondWithJson(w, http.StatusOK, results)
}

func (cfg *apiConfig) handlerSearchGoogleBooks(params metadata.SearchParams, w http.ResponseWriter, r *http.Request) {

	if cfg.googleBooksApiKey == "" {
		respondWithError(w, http.StatusInternalServerError, "missing google books api key in backend setup", fmt.Errorf("missing google books api key in backend setup"))
		return
	}

	results, err := metadata.SearchGoogleBooks(params, cfg.googleBooksApiKey)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "something went wrong while querying google books", err)
		return
	}

	respondWithJson(w, http.StatusOK, results)
}
