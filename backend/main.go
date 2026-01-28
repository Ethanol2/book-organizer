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
	if len(os.Args) > 1 && os.Args[1] == "-r" {
		log.Println("Reset flag")
		dbReset = true
	}

	config, err := initConfig(dbReset)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Environment variables and database loaded successfully")

	mux := http.NewServeMux()
	fHandler := http.FileServer(http.Dir(config.frontendPath))
	mux.Handle("/", fHandler)

	// Downloads Endpoints
	mux.HandleFunc("GET /api/downloads", config.handlerGetDownloads)

	srv := &http.Server{
		Addr:    ":" + config.port,
		Handler: mux,
	}

	scanner := fileScanner.CreateNew(time.Second*5, config.downloadsPath)
	err = scanner.Start(context.Background(), &config.db)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("File scanning started")
	log.Println("Starting server")

	log.Printf("Serving on: http://localhost:%s/\n", config.port)
	log.Fatal(srv.ListenAndServe())
}

func initConfig(dbReset bool) (*apiConfig, error) {

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

	return &apiConfig{
		db:            db,
		frontendPath:  fPath,
		downloadsPath: dPath,
		libraryPath:   lPath,
		port:          port,
	}, nil
}
