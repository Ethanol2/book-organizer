package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"path"

	"github.com/Ethanol2/book-organizer/internal/database"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
)

func (cfg *apiConfig) handlerGetBook(id uuid.UUID, w http.ResponseWriter, r *http.Request) {

	book, err := cfg.db.GetBook(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get book", err)
		return
	}

	respondWithJson(w, http.StatusOK, book)
}

func (cfg *apiConfig) handlerGetBooks(w http.ResponseWriter, r *http.Request) {

	books, err := cfg.db.GetBooks()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get books", err)
		return
	}

	respondWithJson(w, http.StatusOK, books)
}

func (cfg *apiConfig) handlerPostBook(w http.ResponseWriter, r *http.Request) {

	var bookParams database.BookParams
	err := json.NewDecoder(r.Body).Decode(&bookParams)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't read body", err)
		return
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

	respondWithJson(w, http.StatusOK, book)
}

func (cfg *apiConfig) handlerUpdateBook(id uuid.UUID, w http.ResponseWriter, r *http.Request) {

	var update database.BookParams
	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to read body", err)
		return
	}

	book, err := cfg.db.UpdateBook(id, update)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update database", err)
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
	if book.Files.Directory == nil {
		respondWithError(w, http.StatusNoContent, "Book doesn't have a cover associated", err)
		return
	}

	author, series, err := cfg.db.GetPrimaryAuthorAndSeries(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error", err)
		return
	}

	coverPath := path.Join(cfg.libraryPath, author, series, *book.Files.Directory, *book.Files.Cover)
	log.Println("Serving book cover from", coverPath)

	http.ServeFile(w, r, coverPath)
}
