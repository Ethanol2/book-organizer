package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Ethanol2/book-organizer/internal/database"
)

func getCategoryType(r *http.Request) (database.CategoryType, error) {
	catStr := r.PathValue("categoryType")
	var catType database.CategoryType

	switch catStr {

	case "series":
		catType = database.Series
	case "genres":
		catType = database.Genres
	case "narrators":
		catType = database.Narrators
	case "authors":
		catType = database.Authors

	default:
		return database.NoType, fmt.Errorf("Unknown category type")
	}
	return catType, nil
}

func (cfg *apiConfig) handlerPutCategory(w http.ResponseWriter, r *http.Request) {

	var newCat struct {
		Value string `json:"value"`
	}

	err := json.NewDecoder(r.Body).Decode(&newCat)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, BodyDecodeError, err)
	}

	catType, err := getCategoryType(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unknown category type", err)
		return
	}

	var category database.Category
	err = cfg.db.HandleTransaction(func(c *database.Client) error {
		category, err = cfg.db.AddCategory(catType, newCat.Value)
		return err
	})
	if err != nil {
		log.Println(err)
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	respondWithJson(w, http.StatusOK, category)
}

func (cfg *apiConfig) handlerGetAllOfCategory(w http.ResponseWriter, r *http.Request) {

	catType, err := getCategoryType(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unknown category type", err)
		return
	}

	category, err := cfg.db.GetAllOfCategory(catType)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	respondWithJson(w, http.StatusOK, struct {
		Values []database.Category `json:"values"`
	}{category})
}
