package main

import (
	"net/http"

	"github.com/google/uuid"
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
