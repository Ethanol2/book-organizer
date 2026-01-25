package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type apiConfig struct {
	db            *sql.DB
	frontendPath  string
	downloadsPath string
	libraryPath   string
	port          string
}

func main() {

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		log.Fatal("DB_PATH must be set")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("couldn't open database: %v", err)
	}
	defer db.Close()

	fPath := os.Getenv("FRONTENT_PATH")
	if fPath == "" {
		log.Fatal("FRONTENT_PATH must be set")
	}

	dPath := os.Getenv("DOWNLOADS_PATH must be set")
	if dPath == "" {
		log.Fatal("DOWNLOADS_PATH must be set")
	}

	lPath := os.Getenv("LIBRARY_PATH")
	if lPath == "" {
		log.Fatal("LIBRARY_PATH must be set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT must be set")
	}

	config := apiConfig{
		db:            db,
		frontendPath:  fPath,
		downloadsPath: dPath,
		libraryPath:   lPath,
		port:          port,
	}

	mux := http.NewServeMux()
	fHandler := http.FileServer(http.Dir(config.frontendPath))
	mux.Handle("/", fHandler)

	srv := &http.Server{
		Addr:    ":" + config.port,
		Handler: mux,
	}

	log.Printf("Serving on: http://localhost:%s/\n", config.port)
	log.Fatal(srv.ListenAndServe())
}
