package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Ethanol2/book-organizer/internal/database"
	"github.com/Ethanol2/book-organizer/internal/fileScanner"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

type apiConfig struct {
	db            database.Client
	frontendPath  string
	downloadsPath string
	libraryPath   string
	port          string
}

func main() {

	log.Println("Starting book organizer")

	dbReset := false
	dbTestData := false

	for i, _ := range os.Args[1:] {

		switch os.Args[i] {
		case "-r":
			log.Println("Reset flag (-r)")
			dbReset = true

		case "-t":
			log.Println("Test Data Insertion flag (-t)")
			dbTestData = true
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
	mux.HandleFunc("GET /api/downloads", cfg.handlerGetDownloads)
	mux.HandleFunc("GET /api/downloads/{id}", cfg.handlerGetDownload)

	// Category Endpoints
	mux.HandleFunc("POST /api/categories/{categoryType}", cfg.handlerPutCategory)
	mux.HandleFunc("GET /api/categories/{categoryType}", cfg.handlerGetAllOfCategory)

	srv := &http.Server{
		Addr:    ":" + cfg.port,
		Handler: mux,
	}

	scanner := fileScanner.CreateNew(time.Second*5, cfg.downloadsPath)
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

	if insertTestData {
		err = db.InsertTestData()
		if err != nil {
			return nil, err
		}
	}

	return &apiConfig{
		db:            db,
		frontendPath:  fPath,
		downloadsPath: dPath,
		libraryPath:   lPath,
		port:          port,
	}, nil
}
