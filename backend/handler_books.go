package main

import (
	"encoding/json"
	"errors"
	"net/http"

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

	var bookParams database.CreateBookParams
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
