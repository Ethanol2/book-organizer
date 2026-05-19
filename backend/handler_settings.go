package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	// User editable
	GoogleApiKey  = "GOOGLE_BOOKS_API_KEY"
	Port          = "PORT"
	ScanFrequency = "DOWNLOADS_SCAN_FREQUENCY"

	// Hidden
	DatabasePath  = "DB_PATH"
	FrontendPath  = "FRONTEND_PATH"
	DownloadsPath = "DOWNLOADS_PATH"
	LibraryPath   = "LIBRARY_PATH"
	TestDataPath  = "TEST_DATA_PATH"
	TokenSecret   = "TOKEN_SECRET"
)

type Settings struct {
	Port          *string `json:"port,omitempty"`
	GoogleApiKey  *string `json:"google_api_key,omitempty"`
	ScanFrequency *string `json:"scan_frequency,omitempty"`
}

func (cfg *apiConfig) handlerGetSettings(w http.ResponseWriter, r *http.Request) {

	settings := Settings{}

	port := os.Getenv(Port)
	settings.Port = &port

	googleKey := os.Getenv(GoogleApiKey)
	settings.GoogleApiKey = &googleKey

	scanFrequency := os.Getenv(ScanFrequency)
	settings.ScanFrequency = &scanFrequency

	respondWithJson(w, http.StatusOK, settings)
}

func (cfg *apiConfig) handlerPostSettings(w http.ResponseWriter, r *http.Request) {

	var settings Settings
	err := json.NewDecoder(r.Body).Decode(&settings)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, BodyDecodeError, err)
		return
	}

	if settings.ScanFrequency != nil {

		frequency, err := strconv.Atoi(*settings.ScanFrequency)
		if err != nil || frequency < 1 {
			respondWithError(w, http.StatusBadRequest, "scan frequency must be a positive integer greater than 0", err)
			return
		}

		cfg.scanner.Frequency = time.Second * time.Duration(frequency)
		cfg.Env[ScanFrequency] = fmt.Sprint(cfg.scanner.Frequency.Seconds())
	}

	if settings.Port != nil {
		cfg.Env[Port] = *settings.Port
	}

	if settings.GoogleApiKey != nil {
		cfg.Env[GoogleApiKey] = *settings.GoogleApiKey
	}

	err = godotenv.Write(cfg.Env, "./.env")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to write settings to disk", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Println("Updated user settings")
}
