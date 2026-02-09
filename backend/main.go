package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Ethanol2/book-organizer/internal/database"
	"github.com/Ethanol2/book-organizer/internal/fileManagement"
	"github.com/Ethanol2/book-organizer/internal/metadata"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

type apiConfig struct {
	db            database.Client
	frontendPath  string
	downloadsPath string
	downloadsName string
	libraryPath   string
	libraryName   string
	port          string

	googleBooksApiKey string
}

func main() {

	log.Println("Starting book organizer")

	dbReset := false
	dbTestData := false

	for _, arg := range os.Args[1:] {

		switch arg {
		case "-r":
			log.Println("Reset flag (-r)")
			dbReset = true

		case "-t":
			log.Println("Test Data Insertion flag (Resets db) (-t)")
			dbTestData = true
			dbReset = true
		}

	}

	cfg, err := initConfig(dbReset, dbTestData)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Environment variables and database loaded successfully")

	mux := http.NewServeMux()
	fHandler := http.FileServer(http.Dir(cfg.frontendPath))
	mux.Handle("/", fHandler)

	// Downloads Endpoints
	mux.HandleFunc("POST /api/downloads/{id}/associate", uuidMiddleware(cfg.handlerAssociateDownloadToBook))
	mux.HandleFunc("GET /api/downloads", cfg.handlerGetDownloads)
	mux.HandleFunc("GET /api/downloads/{id}", uuidMiddleware(cfg.handlerGetDownload))
	mux.HandleFunc("GET /api/downloads/{id}/cover", uuidMiddleware(cfg.handlerGetDownloadCover))

	// Category Endpoints
	mux.HandleFunc("POST /api/categories/{categoryType}", cfg.handlerPutCategory)
	mux.HandleFunc("GET /api/categories/{categoryType}", cfg.handlerGetAllOfCategory)

	// Book Endpoints
	mux.HandleFunc("POST /api/books", cfg.handlerPostBook)
	mux.HandleFunc("GET /api/books", cfg.handlerGetBooks)
	mux.HandleFunc("GET /api/books/{id}", uuidMiddleware(cfg.handlerGetBook))
	mux.HandleFunc("GET /api/books/{id}/cover", uuidMiddleware(cfg.handlerGetBookCover))
	mux.HandleFunc("PATCH /api/books/{id}", uuidMiddleware(cfg.handlerUpdateBook))

	// Metadata
	mux.HandleFunc("GET /api/metadata/openlibrary", metadataSearchMiddleware(cfg.handlerSearchOpenLibrary))
	mux.HandleFunc("GET /api/metadata/googlebooks", metadataSearchMiddleware(cfg.handlerSearchGoogleBooks))

	// Media
	mux.Handle("/media/downloads/", http.StripPrefix("/media/downloads/", http.FileServer(http.Dir(cfg.downloadsPath))))
	mux.Handle("/media/library/", http.StripPrefix("/media/library/", http.FileServer(http.Dir(cfg.libraryPath))))

	srv := &http.Server{
		Addr:    ":" + cfg.port,
		Handler: mux,
	}

	scanner := fileManagement.CreateNew(time.Second*5, cfg.downloadsPath)
	err = scanner.Start(context.Background(), &cfg.db)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("File scanning started")
	log.Println("Starting server")

	log.Printf("Serving on: http://localhost:%s/\n", cfg.port)
	log.Fatal(srv.ListenAndServe())
}

func initConfig(dbReset, insertTestData bool) (*apiConfig, error) {

	godotenv.Load(".env")

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		return nil, fmt.Errorf("DB_PATH must be set")
	}

	if dbReset {
		err := os.Remove(dbPath)
		if err != nil {
			log.Println(err)
		}
	}

	db, err := database.NewClient(dbPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't open database: %v", err)
	}

	fPath := os.Getenv("FRONTENT_PATH")
	if fPath == "" {
		return nil, fmt.Errorf("FRONTENT_PATH must be set")
	}

	dPath := os.Getenv("DOWNLOADS_PATH")
	if dPath == "" {
		return nil, fmt.Errorf("DOWNLOADS_PATH must be set")
	}

	lPath := os.Getenv("LIBRARY_PATH")
	if lPath == "" {
		return nil, fmt.Errorf("LIBRARY_PATH must be set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		return nil, fmt.Errorf("PORT must be set")
	}

	gbApiKey := os.Getenv("GOOGLE_BOOKS_API_KEY")
	if gbApiKey == "" {
		log.Println("no google books api key in env variables. Google books search won't work")
	}

	if insertTestData {
		err = db.InsertTestData()
		if err != nil {
			return nil, err
		}
	}

	return &apiConfig{
		db:                db,
		frontendPath:      fPath,
		downloadsPath:     dPath,
		downloadsName:     "/media/downloads",
		libraryPath:       lPath,
		libraryName:       "/media/library",
		port:              port,
		googleBooksApiKey: gbApiKey,
	}, nil
}

func uuidMiddleware(handler func(uuid.UUID, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		idStr := r.PathValue("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid id", err)
			return
		}

		handler(id, w, r)
	}

}

func metadataSearchMiddleware(handler func(metadata.SearchParams, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		var searchParams metadata.SearchParams
		err := json.NewDecoder(r.Body).Decode(&searchParams)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "couldn't read body", err)
			return
		}

		handler(searchParams, w, r)
	}
}
