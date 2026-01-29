package main

import (
	"log"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerGetDownloads(w http.ResponseWriter, r *http.Request) {

	log.Println("Fetching downloads")
	downloads, err := cfg.db.GetDownloads()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "something went wrong retrieving downloads", err)
	}
	respondWithJson(w, http.StatusOK, downloads)
}

func (cfg *apiConfig) handlerGetDownload(w http.ResponseWriter, r *http.Request) {

	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid video ID", err)
		return
	}

	download, err := cfg.db.GetDownload(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get download", err)
		return
	}

	respondWithJson(w, http.StatusOK, download)
}
