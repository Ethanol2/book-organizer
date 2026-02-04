package main

import (
	"encoding/json"
	"net/http"

	"github.com/Ethanol2/book-organizer/metadata"
)

func (cfg *apiConfig) handlerSearchOpenLibrary(w http.ResponseWriter, r *http.Request) {

	var searchParams metadata.OpenLibrarySearchParams
	err := json.NewDecoder(r.Body).Decode(&searchParams)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't read body", err)
		return
	}

	results, err := metadata.SearchOpenLibrary(searchParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "something went wrong while querying openlibrary", err)
		return
	}

	respondWithJson(w, http.StatusOK, results)
}
