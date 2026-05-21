package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/Ethanol2/book-organizer/internal/env"
	"github.com/Ethanol2/book-organizer/internal/logger"
)

type Settings struct {
	Port          *string `json:"port,omitempty"`
	GoogleApiKey  *string `json:"google_api_key,omitempty"`
	ScanFrequency *string `json:"scan_frequency,omitempty"`
}

// #region Settings Handlers

func (cfg *apiConfig) handlerGetSettings(w http.ResponseWriter, r *http.Request) {

	settings := Settings{}
	values := cfg.Env.GetMultiple(env.Port, env.GoogleApiKey, env.ScanFrequency)

	port := values[env.Port]
	settings.Port = &port

	googleKey := values[env.GoogleApiKey]
	settings.GoogleApiKey = &googleKey

	scanFrequency := values[env.ScanFrequency]
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

	update := map[string]string{}

	if settings.ScanFrequency != nil {

		frequency, err := strconv.Atoi(*settings.ScanFrequency)
		if err != nil || frequency < 1 {
			respondWithError(w, http.StatusBadRequest, "scan frequency must be a positive integer greater than 0", err)
			return
		}

		cfg.scanner.Frequency = time.Second * time.Duration(frequency)
		update[env.ScanFrequency] = fmt.Sprint(cfg.scanner.Frequency.Seconds())
	}

	if settings.Port != nil {
		update[env.Port] = *settings.Port
	}

	if settings.GoogleApiKey != nil {
		update[env.GoogleApiKey] = *settings.GoogleApiKey
	}

	err = cfg.Env.SetMultiple(update)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to write settings to disk", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Println("Updated user settings")
}

func (cfg *apiConfig) handlerGetLogs(w http.ResponseWriter, r *http.Request) {

	var err error

	offset := 0
	if str := r.URL.Query().Get("offset"); str != "" {
		offset, err = strconv.Atoi(str)
		if err != nil || offset < 0 {
			respondWithError(w, http.StatusBadRequest, "Offset must be an integer greater or equal to 0", err)
			return
		}
	}
	count := 20
	if str := r.URL.Query().Get("count"); str != "" {
		count, err = strconv.Atoi(str)
		if err != nil || count <= 0 {
			respondWithError(w, http.StatusBadRequest, "Count must an integer greater or equal to 1", err)
			return
		}
	}

	logs, err := logger.GetLogs(path.Join(cfg.logsPath, "logs.txt"), count, offset)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, FileOpenError, err)
		return
	}

	respondWithJson(w, http.StatusOK, struct {
		Logs []logger.Log `json:"logs"`
	}{Logs: logs})
}

func (cfg *apiConfig) handlerShutdown(shutdownChan chan os.Signal) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		_, err := authenticate(cfg.authRequired, r, cfg.Env.Get(env.TokenSecret))
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, AuthBadAuthorization, err)
			return
		}

		w.WriteHeader(http.StatusOK)

		log.Println("Shutdown command received")

		close(shutdownChan)
	}

}
