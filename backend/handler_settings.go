package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/Ethanol2/book-organizer/internal/auth"
	"github.com/Ethanol2/book-organizer/internal/logger"
	"github.com/joho/godotenv"
)

const (
	EnvPath = "./.env"

	// User editable
	GoogleApiKey  = "GOOGLE_BOOKS_API_KEY"
	Port          = "PORT"
	ScanFrequency = "DOWNLOADS_SCAN_FREQUENCY"

	// Hidden
	DataPath      = "DATA_PATH"
	FrontendPath  = "FRONTEND_PATH"
	DownloadsPath = "DOWNLOADS_PATH"
	LibraryPath   = "LIBRARY_PATH"
	TokenSecret   = "TOKEN_SECRET"
)

type env struct {
	mu  *sync.Mutex
	env map[string]string
}

type Settings struct {
	Port          *string `json:"port,omitempty"`
	GoogleApiKey  *string `json:"google_api_key,omitempty"`
	ScanFrequency *string `json:"scan_frequency,omitempty"`
}

// #region Settings Handlers

func (cfg *apiConfig) handlerGetSettings(w http.ResponseWriter, r *http.Request) {

	settings := Settings{}
	values := cfg.Env.GetMultiple(Port, GoogleApiKey, ScanFrequency)

	port := values[Port]
	settings.Port = &port

	googleKey := values[GoogleApiKey]
	settings.GoogleApiKey = &googleKey

	scanFrequency := values[ScanFrequency]
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
		update[ScanFrequency] = fmt.Sprint(cfg.scanner.Frequency.Seconds())
	}

	if settings.Port != nil {
		update[Port] = *settings.Port
	}

	if settings.GoogleApiKey != nil {
		update[GoogleApiKey] = *settings.GoogleApiKey
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

// #region Env Functions

func initEnv() (env, error) {
	values, err := godotenv.Read(EnvPath)
	if err != nil {
		return env{}, err
	}

	needsUpdate := false

	dataPath := ""
	if path, ok := values[DataPath]; !ok || path == "" {
		fmt.Println("No data path defined in the .env. Setting value to \"./data/\"")
		dataPath = "./data/"
		values[DataPath] = dataPath
		needsUpdate = true
	} else {
		dataPath = path
	}

	if fPath, ok := values[FrontendPath]; !ok || fPath == "" {
		fmt.Println("No frontend path defined in the .env. Setting value to \"./static/\"")
		values[FrontendPath] = "./static/"
		needsUpdate = true
	}

	if dPath, ok := values[DownloadsPath]; !ok || dPath == "" {
		fmt.Println("No downloads path defined in the .env. Setting value to \"./downloads/\"")
		values[DownloadsPath] = "./downloads/"
		needsUpdate = true
	}

	if lPath, ok := values[LibraryPath]; !ok || lPath == "" {
		fmt.Println("No library path defined in the .env. Setting value to \"./library\"")
		values[LibraryPath] = "./library/"
		needsUpdate = true
	}

	if port, ok := values[Port]; !ok || port == "" {
		fmt.Println("No port defined in the .env. Setting value to 8080")
		values[Port] = "8080"
		needsUpdate = true
	}

	if key, ok := values[GoogleApiKey]; !ok || key == "" {
		fmt.Println("No google books api key defined in the .env. Google Books will not work without it. This can also be set via the settings")
	}

	if secret, ok := values[TokenSecret]; !ok || secret == "" {
		fmt.Println("No secret key defined in the .env. Applying a newly generated key.")
		values[TokenSecret] = auth.MakeToken()
		needsUpdate = true
	}

	if fStr, ok := values[ScanFrequency]; !ok || fStr == "" {
		fmt.Println("No scan frequency was defined in the .env. Setting value to 5 seconds")
		values[ScanFrequency] = "5"
		needsUpdate = true
	} else {
		seconds, err := strconv.Atoi(fStr)
		if err != nil || seconds < 1 {
			log.Fatal("scan frequency must be a positive integer greater than 0")
		}
	}

	if needsUpdate {
		err = godotenv.Write(values, EnvPath)
		if err != nil {
			return env{}, err
		}
	}

	return env{
		&sync.Mutex{},
		values,
	}, nil
}

func (env *env) Get(key string) string {
	(*env.mu).Lock()
	defer (*env.mu).Unlock()
	return env.env[key]
}

func (env *env) GetInt(key string) int {
	str := env.Get(key)
	if str == "" {
		log.Fatalf("Number \"%s\" requested from the env variables, but nothing exists", key)
	}

	num, err := strconv.Atoi(str)
	if err != nil {
		log.Fatalf("Number \"%s\" requested from the env variables, but it's an invalid value: %v", key, err)
	}

	return num
}

func (env *env) GetMultiple(keys ...string) map[string]string {
	(*env.mu).Lock()
	defer (*env.mu).Unlock()

	values := map[string]string{}
	for _, key := range keys {
		values[key] = env.env[key]
	}

	return values
}

func (env *env) Set(key, value string) error {
	(*env.mu).Lock()
	defer (*env.mu).Unlock()

	env.env[key] = value
	return godotenv.Write(env.env, EnvPath)
}

func (env *env) SetMultiple(values map[string]string) error {
	(*env.mu).Lock()
	defer (*env.mu).Unlock()

	for key, value := range values {
		env.env[key] = value
	}

	return godotenv.Write(env.env, EnvPath)
}
