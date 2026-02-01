package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/Ethanol2/book-organizer/internal/database"
	"github.com/Ethanol2/book-organizer/internal/fileManagement"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
)

func (cfg *apiConfig) handlerGetDownloads(w http.ResponseWriter, r *http.Request) {

	log.Println("Fetching downloads")
	downloads, err := cfg.db.GetDownloads()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "something went wrong retrieving downloads", err)
	}
	respondWithJson(w, http.StatusOK, downloads)
}

func (cfg *apiConfig) handlerGetDownload(id uuid.UUID, w http.ResponseWriter, r *http.Request) {

	download, err := cfg.db.GetDownload(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get download", err)
		return
	}

	respondWithJson(w, http.StatusOK, download)
}

func (cfg *apiConfig) handlerAssociateDownloadToBook(downloadId uuid.UUID, w http.ResponseWriter, r *http.Request) {

	downloadDir, err := cfg.db.GetDownloadDir(downloadId)
	if err != nil {
		if errors.Is(err, sqlite3.ErrNotFound) {
			respondWithError(w, http.StatusBadRequest, "Download not found", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Database error", err)
		return
	}

	var bookIdStruct struct {
		BookId uuid.UUID `json:"book_id"`
	}
	err = json.NewDecoder(r.Body).Decode(&bookIdStruct)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't read body", err)
		return
	}

	bookExists, err := cfg.db.CheckBookExists(downloadId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error", err)
		return
	}
	if !bookExists {
		respondWithError(w, http.StatusBadRequest, "Book not found", err)
		return
	}

	authorDir := "Unknown"
	authors, err := cfg.db.GetCategoryTypesAssociatedWithBook(bookIdStruct.BookId.String(), database.Authors)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error", err)
		return
	}
	if len(authors) > 0 {
		authorDir = authors[0].Name
	}

	seriesDir := ""
	series, err := cfg.db.GetCategoryTypesAssociatedWithBook(bookIdStruct.BookId.String(), database.Series)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error", err)
		return
	}
	if len(series) > 0 {
		seriesDir = series[0].Name
	}

	oldPath, newPath, err := fileManagement.MoveFiles(downloadDir, cfg.libraryPath, cfg.downloadsPath, authorDir, seriesDir)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong moving files", err)
		return
	}

	book, err := cfg.db.AssociateBookAndDownload(bookIdStruct.BookId, downloadId)
	if err != nil {
		err = fileManagement.MoveFilesWithPaths(newPath, oldPath)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to associate the book and files, and failed to move files back from the library to downloads", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to associate the book and files. Files have been returned to downloads", err)
		return
	}

	respondWithJson(w, http.StatusOK, book)
}
