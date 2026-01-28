package main

import (
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerGetDownloads(w http.ResponseWriter, r *http.Request) {

	log.Println("Fetching downloads")
	downloads, err := cfg.db.GetDownloads()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "something went wrong retrieving downloads", err)
	}
	respondWithJson(w, http.StatusOK, downloads)
}
