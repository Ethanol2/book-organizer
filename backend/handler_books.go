package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/Ethanol2/book-organizer/internal/database"
	"github.com/Ethanol2/book-organizer/internal/fileManagement"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
)

func (cfg *apiConfig) handlerGetBook(id uuid.UUID, w http.ResponseWriter, r *http.Request) {

	book, err := cfg.db.GetBook(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get book", err)
		return
	}

	if book.Files.Root == nil {
		book.Files.Prepend(cfg.libraryName)
	}

	respondWithJson(w, http.StatusOK, book)
}

func (cfg *apiConfig) handlerGetBooks(w http.ResponseWriter, r *http.Request) {

	books, err := cfg.db.GetBooks()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get books", err)
		return
	}

	for i := range books {
		books[i].Files.Prepend(cfg.libraryName)
	}

	log.Println("Fetching books")

	respondWithJson(w, http.StatusOK, books)
}

func (cfg *apiConfig) handlerPostBook(w http.ResponseWriter, r *http.Request) {

	var bookParams database.BookParams
	err := json.NewDecoder(r.Body).Decode(&bookParams)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't read body", err)
		return
	}

	var coverFile *os.File
	if bookParams.Cover != nil {
		coverFile, err = fileManagement.DownloadTempFile(*bookParams.Cover)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "failed to fetch cover from url", err)
			return
		}
		defer coverFile.Close()
	}

	book, err := cfg.db.AddBook(bookParams)
	if err != nil {

		if errors.Is(err, sqlite3.ErrConstraintUnique) {
			respondWithError(w, http.StatusBadRequest, "Books can't share ISBN or ASIN numbers to prevent duplicates", err)
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Couldn't add book to db", err)
		return
	}

	if coverFile != nil {
		err = fileManagement.MoveFilesWithPaths(coverFile.Name(), path.Join(cfg.metadataPath, book.Id.String()+path.Ext(coverFile.Name())))
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "failed to create cover file", err)
			return
		}
	}

	respondWithJson(w, http.StatusOK, book)
}

func (cfg *apiConfig) handlerUpdateBook(id uuid.UUID, w http.ResponseWriter, r *http.Request) {

	var bookParams database.BookParams
	err := json.NewDecoder(r.Body).Decode(&bookParams)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to read body", err)
		return
	}

	book, err := cfg.db.UpdateBook(id, bookParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update database", err)
		return
	}

	book.Files.Prepend(cfg.libraryName)
	respondWithJson(w, http.StatusOK, book)
}

func (cfg *apiConfig) handlerUpdateBookCover(id uuid.UUID, w http.ResponseWriter, r *http.Request) {

	tmp, err := fileManagement.CreateTempFileFromRequest(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to download image", err)
		return
	}
	defer tmp.Close()

	cfg.db.Begin()
	defer cfg.db.Rollback()

	oldPath, newPath, err := cfg.db.UpdateBookCover(id, path.Ext(tmp.Name()))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error", err)
		return
	}

	if oldPath == "" {
		oldPath = path.Join(cfg.metadataPath, id.String()+path.Ext(tmp.Name()))
	}

	err = os.Remove(oldPath)
	if err != nil {
		log.Println("Failed to remove old cover")
		respondWithError(w, http.StatusInternalServerError, "File error", err)
		return
	}

	if newPath == "" {
		newPath = path.Join(cfg.metadataPath, id.String()+path.Ext(tmp.Name()))
	}

	err = fileManagement.MoveFilesWithPaths(tmp.Name(), newPath)
	if err != nil {
		log.Println("Failed to move new cover to path")
		respondWithError(w, http.StatusInternalServerError, "File error", err)
		return
	}

	err = cfg.db.Commit()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error", err)
		return
	}

	book, err := cfg.db.GetBook(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error", err)
		return
	}

	respondWithJson(w, http.StatusOK, book)

}

func (cfg *apiConfig) handlerGetBookCover(id uuid.UUID, w http.ResponseWriter, r *http.Request) {

	book, err := cfg.db.GetBook(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error", err)
		return
	}
	if book.Id == nil {
		respondWithError(w, http.StatusBadRequest, "Book not found", err)
		return
	}
	if book.Files.Cover == nil {
		respondWithError(w, http.StatusNoContent, "Book doesn't have a cover associated", err)
		return
	}

	coverPath := ""
	if book.Files.Root == nil {
		coverPath = path.Join(cfg.metadataPath, *book.Files.Cover)
	} else {
		author, series, err := cfg.db.GetPrimaryAuthorAndSeries(id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Database error", err)
			return
		}

		coverPath = path.Join(cfg.libraryPath, author, series, *book.Files.Cover)
	}
	log.Println("Serving book cover from", coverPath)

	http.ServeFile(w, r, coverPath)
}
