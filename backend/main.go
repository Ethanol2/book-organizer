package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Ethanol2/book-organizer/internal/cache"
	"github.com/Ethanol2/book-organizer/internal/database"
	"github.com/Ethanol2/book-organizer/internal/fileManagement"
	"github.com/Ethanol2/book-organizer/internal/metadata"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

const helpMsg string = `

Book Organizer Backend
An app written by Ethan Colucci for organizing your digital book library.

Project Repo	https://github.com/Ethanol2/book-organizer

My Website:		https://ethanasc.ca
My GitHub:		https://github.com/Ethanol2
My LinkedIn:	https://www.linkedin.com/in/ethan-colucci/

[-h, --help]:
	Displays this message and exits. All other commands run the app.

[-r, --reset]:
	Deletes the appData.db file, reseting the database, and empties the metadata folder.

[-l, --clear-library]:
	Clears the contents of the library, defined in the .env as LIBRARY_PATH

[-d, --clear-downloads]:
	Clears the contents of the downloads, defined in the .env as DOWNLOAD_PATH

[-t, --test-dataset] <l, test-library> <d, test-downloads>:
	Deletes the appData.db file, reseting the database, empties the metadata folder, 
	and adds a test dataset defined in ./backend/internal/database/testdata.go

	<l, test-library>:
		Clears the library and creates a test library structure using the test dataset.
		Inserts 15 books with random text and audio files. All books will have a cover.
		Half will have a metadata file.
	
	<d, test-downloads>:
		Clears the downloads and creates a test download structure using the test dataset.
		Inserts 10 total downloads: 
			Count | Has Cover | Has Metadata | Audio Files | Text Files
			-----------------------------------------------------------
			  1   |    true   |     false    |      1      |     1
			  1   |    false  |     false    |      1      |     1
			  1   |    true   |     false    |      10     |     1
			  1   |    false  |     false    |      1      |     10
			  3   |    true   |     true     |      5      |     1
			  3   |    false  |     true     |      1      |     5

`

type apiConfig struct {
	// System Structs
	db      database.Client
	mdCache cache.Cache

	// Folder Paths
	frontendPath  string
	downloadsPath string
	libraryPath   string
	metadataPath  string
	testDataPath  string

	// Endpoint Paths
	downloadsName string
	libraryName   string

	// Other
	port              string
	googleBooksApiKey string
}

type cliFlags struct {
	dbReset      bool
	dbTestData   bool
	libTestData  bool
	downTestData bool

	clearLibrary   bool
	clearDownloads bool
}

func main() {

	flags, err := getFlags()
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Println("Starting book organizer")

	cfg, err := initConfig(flags)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Environment variables and database loaded successfully")

	mux := http.NewServeMux()
	fHandler := http.FileServer(http.Dir(cfg.frontendPath))
	mux.Handle("/", fHandler)

	// Library Endpoints
	mux.HandleFunc("GET /api/library/scan", cfg.handlerGetScanLibrary)

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
	mux.HandleFunc("PATCH /api/books/{id}/cover", uuidMiddleware(cfg.handlerUpdateBookCover))
	mux.HandleFunc("DELETE /api/books/{id}", uuidMiddleware(cfg.handlerDeleteBook))

	// Metadata
	mux.HandleFunc("GET /api/metadata/", cfg.handlerMetadataSearch)
	mux.HandleFunc("GET /api/metadata/{id}", cfg.handlerGetMetadataBookDetails)

	// Media
	mux.Handle("/media/downloads/", http.StripPrefix("/media/downloads/", http.FileServer(http.Dir(cfg.downloadsPath))))
	mux.Handle("/media/library/", http.StripPrefix("/media/library/", http.FileServer(http.Dir(cfg.libraryPath))))
	mux.Handle("/media/metadata/", http.StripPrefix("/media/metadata/", http.FileServer(http.Dir(cfg.metadataPath))))

	srv := &http.Server{
		Addr:    ":" + cfg.port,
		Handler: mux,
	}

	scanner := fileManagement.Scanner{
		Frequency: time.Second * 5,
		Directory: cfg.downloadsPath,

		AddHandler:    cfg.db.AddDownloads,
		UpdateHandler: cfg.db.UpdateDownloadFiles,
		DeleteHandler: cfg.db.DeleteDownload,
		GetExisting:   cfg.db.GetAllDownloadsIdsAndDirs,
	}
	err = scanner.Start(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("File scanning started")
	log.Println("Starting server")

	log.Printf("Serving on: http://localhost:%s/\n", cfg.port)
	log.Fatal(srv.ListenAndServe())
}

func getFlags() (cliFlags, error) {

	flags := cliFlags{}
	tArgs := false

	for _, arg := range os.Args[1:] {

		arg = strings.ToLower(arg)

		if arg == "--help" || arg == "-h" {
			return cliFlags{}, fmt.Errorf(helpMsg)
		}

		if tArgs {
			switch arg {
			case "l", "test-library":
				fmt.Println("Test library creation arg (-t l)")
				flags.libTestData = true
				continue

			case "d", "test-downloads":
				fmt.Println("Test downloads creation arg (-t d)")
				flags.downTestData = true
				continue
			}
		}

		switch arg {
		case "--reset", "-r":
			fmt.Println("Reset flag (-r)")
			flags.dbReset = true
			tArgs = false

		case "-l", "--clear-library":
			fmt.Println("Clear library flag (-l)")
			flags.clearLibrary = true
			tArgs = false

		case "-d", "--clear-downloads":
			fmt.Println("Clear downloads flag (-d)")
			flags.clearDownloads = true
			tArgs = false

		case "--test-dataset", "-t":
			fmt.Println("Test Data Insertion flag (Resets db) (-t)")
			flags.dbReset = true
			flags.dbTestData = true
			tArgs = true

		default:
			return cliFlags{}, fmt.Errorf("unknown command. Use -h for a list of commands")
		}
	}

	fmt.Println()

	return flags, nil
}

func initConfig(flags cliFlags) (*apiConfig, error) {

	godotenv.Load(".env")

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		return nil, fmt.Errorf("DB_PATH must be set")
	}

	metadataPath := "./data/metadata"

	if flags.dbReset {

		// Remove the database file
		err := os.Remove(dbPath)
		if err != nil {
			fmt.Println(err)
		}

		// Clear the metadata folder
		if _, err := os.Stat(metadataPath); err == nil {
			files, err := os.ReadDir(metadataPath)
			if err == nil {
				for _, file := range files {
					if file.IsDir() {
						continue
					}
					err = os.Remove(path.Join(metadataPath, file.Name()))
					if err != nil {
						fmt.Println("Error while trying to delete the file \"", file.Name(), "\" => ", err)
					}
				}
			} else {
				fmt.Println("Error when trying to read the contents of \"", metadataPath, "\" =>", err)
			}
		} else {
			err = os.Mkdir(metadataPath, 0644)
			if err != nil {
				fmt.Println("Error while creating the metadata directory")
				return nil, err
			}
		}
	}

	db, err := database.NewClient(dbPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't open database: %v", err)
	}

	fPath := os.Getenv("FRONTEND_PATH")
	if fPath == "" {
		return nil, fmt.Errorf("FRONTEND_PATH must be set")
	}

	dPath := os.Getenv("DOWNLOADS_PATH")
	if dPath == "" {
		return nil, fmt.Errorf("DOWNLOADS_PATH must be set")
	}

	lPath := os.Getenv("LIBRARY_PATH")
	if lPath == "" {
		return nil, fmt.Errorf("LIBRARY_PATH must be set")
	}

	tPath := os.Getenv("TEST_DATA_PATH")
	if tPath == "" {
		fmt.Println("TEST_DATA_PATH not set. Needed for the -t insert test data flag")
		if flags.dbTestData {
			return nil, fmt.Errorf("insert test data flag triggered but no test folder path was provided")
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		return nil, fmt.Errorf("PORT must be set")
	}

	gbApiKey := os.Getenv("GOOGLE_BOOKS_API_KEY")
	if gbApiKey == "" {
		fmt.Println("no google books api key in env variables. Google books search won't work")
	}

	if flags.clearDownloads {
		err = fileManagement.RemoveDirectoryContents(dPath)
		if err != nil {
			return nil, err
		}
	}

	if flags.clearLibrary {
		fmt.Println("Deleting library")
		err = fileManagement.RemoveDirectoryContents(lPath)
		if err != nil {
			return nil, err
		}
	}

	if flags.dbTestData {

		fmt.Println("Creating test data path")
		err = fileManagement.CreateDirectory(tPath)
		if err != nil {
			return nil, err
		}

		testBooks, err := db.InsertTestData(metadataPath, tPath)
		if err != nil {
			return nil, err
		}

		// Insert Library
		if flags.libTestData {

			// Inserts 15 books with random text and audio files. All books will have a cover.
			// Half will have a metadata file.

			fmt.Println()
			fmt.Println("======= Creating test library =======")

			handleDirs := func(book database.Book) (string, error) {
				bookPath := book.Authors[0].Name

				err = fileManagement.CreateDirectory(path.Join(lPath, bookPath))
				if err != nil {
					return "", err
				}

				bookDir := book.Title
				if len(book.Series) > 0 {

					bookPath = path.Join(bookPath, book.Series[0].Name)

					if book.Series[0].Index != nil {
						bookDir = *book.Series[0].Index + " - " + bookDir
					}

					err = fileManagement.CreateDirectory(path.Join(lPath, bookPath))
					if err != nil {
						return "", err
					}
				}

				bookPath = path.Join(bookPath, bookDir)
				return bookPath, nil
			}

			err = db.Begin()
			if err != nil {
				return nil, err
			}
			defer db.Rollback()

			// No metadata
			for _, book := range testBooks[:7] {

				bookPath, err := handleDirs(book)
				if err != nil {
					return nil, err
				}

				book.Files, err = fileManagement.CreateTestDirectory(
					bookPath,
					lPath,
					nil,
					path.Join(metadataPath, book.Id.String()+".jpg"), rand.Intn(10)+1, rand.Intn(10)+1,
				)

				err = book.ApplyBookFiles(db)
				if err != nil {
					return nil, err
				}
			}

			// With metadata
			for _, book := range testBooks[8:15] {

				bookPath, err := handleDirs(book)
				if err != nil {
					return nil, err
				}

				book.Files, err = fileManagement.CreateTestDirectory(
					bookPath,
					lPath,
					metadata.BookToMetadata(book),
					path.Join(metadataPath, book.Id.String()+".jpg"), rand.Intn(10)+1, rand.Intn(10)+1,
				)

				err = book.ApplyBookFiles(db)
				if err != nil {
					return nil, err
				}
			}

			err = db.Commit()
			if err != nil {
				return nil, err
			}
		}

		// Insert downloads
		if flags.downTestData {
			// Count | Has Cover | Has Metadata | Audio Files | Text Files
			// 	-----------------------------------------------------------
			// 	  1   |    true   |     false    |      1      |     1
			// 	  1   |    false  |     false    |      1      |     1
			// 	  1   |    true   |     false    |      10     |     1
			// 	  1   |    false  |     false    |      1      |     10
			// 	  3   |    true   |     true     |      5      |     1
			// 	  3   |    false  |     true     |      1      |     5

			fmt.Println()
			fmt.Println("======= Creating test downloads =======")

			testDownloads := []struct {
				title, cover string
				md           *fileManagement.MetadataFile
				audio        int
				text         int
			}{
				{
					title: testBooks[0].Title, cover: testBooks[0].Id.String() + ".jpg", md: nil, audio: 1, text: 1,
				}, {
					title: testBooks[1].Title, cover: "", md: nil, audio: 1, text: 1,
				}, {
					title: testBooks[2].Title, cover: testBooks[2].Id.String() + ".jpg", md: nil, audio: 10, text: 1,
				}, {
					title: testBooks[3].Title, cover: "", md: nil, audio: 1, text: 10,
				}, {
					title: testBooks[4].Title, cover: testBooks[4].Id.String() + ".jpg", md: metadata.BookToMetadata(testBooks[4]), audio: 5, text: 1,
				}, {
					title: testBooks[5].Title, cover: testBooks[5].Id.String() + ".jpg", md: metadata.BookToMetadata(testBooks[5]), audio: 5, text: 1,
				}, {
					title: testBooks[6].Title, cover: testBooks[6].Id.String() + ".jpg", md: metadata.BookToMetadata(testBooks[6]), audio: 5, text: 1,
				}, {
					title: testBooks[7].Title, cover: "", md: metadata.BookToMetadata(testBooks[7]), audio: 1, text: 5,
				}, {
					title: testBooks[8].Title, cover: "", md: metadata.BookToMetadata(testBooks[8]), audio: 1, text: 5,
				}, {
					title: testBooks[9].Title, cover: "", md: metadata.BookToMetadata(testBooks[9]), audio: 1, text: 5,
				},
			}

			for _, d := range testDownloads {

				if d.cover != "" {
					d.cover = path.Join(metadataPath, d.cover)
				}

				_, err = fileManagement.CreateTestDirectory(
					path.Join(d.title),
					dPath,
					d.md,
					d.cover,
					d.audio, d.text,
				)
				if err != nil {
					return nil, err
				}
			}

		}

	}

	return &apiConfig{
		db:                db,
		mdCache:           cache.NewCache(time.Minute * 5),
		frontendPath:      fPath,
		downloadsPath:     dPath,
		downloadsName:     "/media/downloads",
		libraryPath:       lPath,
		libraryName:       "/media/library",
		metadataPath:      metadataPath,
		testDataPath:      tPath,
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
