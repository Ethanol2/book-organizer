package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Ethanol2/book-organizer/internal/auth"
	"github.com/Ethanol2/book-organizer/internal/cache"
	"github.com/Ethanol2/book-organizer/internal/database"
	"github.com/Ethanol2/book-organizer/internal/env"
	"github.com/Ethanol2/book-organizer/internal/fileManagement"
	"github.com/Ethanol2/book-organizer/internal/logger"
	"github.com/Ethanol2/book-organizer/internal/metadata"
	"github.com/google/uuid"
)

type apiConfig struct {
	// System Structs
	db      database.Client
	mdCache cache.Cache
	scanner *fileManagement.Scanner

	// Environment Variables
	Env env.Env

	// Folder Paths
	metadataPath string
	logsPath     string

	// Endpoint Paths
	downloadsName string
	libraryName   string

	// Other
	authRequired bool
}

type cliFlags struct {
	dbReset        bool
	booksReset     bool
	downloadsReset bool
	dbTestData     bool
	libTestData    bool
	downTestData   bool

	clearLibrary   bool
	clearDownloads bool

	usersReset    bool
	clearSessions bool
}

func main() {

	flags, err := getFlags(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Starting book organizer")

	cfg, err := createConfig(flags)
	if err != nil {
		fmt.Println("Init Error: " + err.Error())
		os.Exit(1)
	}

	logger.InitLogger(cfg.logsPath)
	log.Println("Environment variables and database loaded successfully")

	mux := http.NewServeMux()

	// Create the channel that handles the shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// General Handlers
	mux.HandleFunc("/", handlerFrontend)
	mux.HandleFunc("/health-check", cfg.handlerHealthCheckfunc)

	// Auth Endpoints
	mux.HandleFunc("GET /api/auth/status", cfg.handlerGetAuthStatus)
	mux.HandleFunc("POST /api/auth/login", cfg.handlerLogin)
	mux.HandleFunc("POST /api/auth/logout", cfg.handlerLogout)
	mux.HandleFunc("POST /api/auth/register", cfg.handlerRegister)
	mux.HandleFunc("DELETE /api/auth/users/{id}", cfg.uuidMiddleware(cfg.handlerDeleteUser))
	mux.HandleFunc("PUT /api/auth/users/{id}/reset-password", cfg.uuidMiddleware(cfg.handlerUpdatePassword))
	mux.HandleFunc("POST /api/auth/refresh", cfg.handlerRefresh)

	// Setting Endpoints
	mux.HandleFunc("GET /api/settings", cfg.authMiddleware(cfg.handlerGetSettings))
	mux.HandleFunc("POST /api/settings", cfg.authMiddleware(cfg.handlerPostSettings))
	mux.HandleFunc("POST /api/settings/shutdown", cfg.handlerShutdown(sigChan))

	// Logs Endpoints
	mux.HandleFunc("GET /api/logs", cfg.authMiddleware(cfg.handlerGetLogs))

	// Downloads Endpoints
	mux.HandleFunc("POST /api/downloads/{id}/associate", cfg.uuidMiddleware(cfg.handlerAssociateDownloadToBook))
	mux.HandleFunc("GET /api/downloads", cfg.authMiddleware(cfg.handlerGetDownloads))
	mux.HandleFunc("GET /api/downloads/{id}", cfg.uuidMiddleware(cfg.handlerGetDownload))
	mux.HandleFunc("GET /api/downloads/{id}/cover", cfg.uuidMiddleware(cfg.handlerGetDownloadCover))
	mux.HandleFunc("GET /api/downloads/{id}/metadata", cfg.uuidMiddleware(cfg.handlerGetDownloadMetadata))

	// Category Endpoints
	mux.HandleFunc("POST /api/categories/{categoryType}", cfg.authMiddleware(cfg.handlerPutCategory))
	mux.HandleFunc("GET /api/categories/{categoryType}", cfg.authMiddleware(cfg.handlerGetAllOfCategory))

	// Book Endpoints
	mux.HandleFunc("GET /api/library/scan", cfg.authMiddleware(cfg.handlerGetScanLibrary))
	mux.HandleFunc("POST /api/books", cfg.authMiddleware(cfg.handlerPostBook))
	mux.HandleFunc("GET /api/books", cfg.authMiddleware(cfg.handlerGetBooks))
	mux.HandleFunc("GET /api/books/{id}", cfg.uuidMiddleware(cfg.handlerGetBook))
	mux.HandleFunc("PATCH /api/books/{id}", cfg.uuidMiddleware(cfg.handlerUpdateBook))
	mux.HandleFunc("DELETE /api/books/{id}", cfg.uuidMiddleware(cfg.handlerDeleteBook))

	// Metadata
	mux.HandleFunc("GET /api/metadata/", cfg.authMiddleware(cfg.handlerMetadataSearch))
	mux.HandleFunc("GET /api/metadata/{id}", cfg.authMiddleware(cfg.handlerGetMetadataBookDetails))

	// Media
	mux.Handle("/media/downloads/", http.StripPrefix(cfg.downloadsName, http.FileServer(http.Dir(cfg.Env.Get(env.DownloadsPath)))))
	mux.Handle("/media/library/", http.StripPrefix(cfg.libraryName, http.FileServer(http.Dir(cfg.Env.Get(env.LibraryPath)))))
	mux.Handle("/media/metadata/", http.StripPrefix("/media/metadata/", http.FileServer(http.Dir(cfg.metadataPath))))

	srv := &http.Server{
		Addr:    ":" + cfg.Env.Get(env.Port),
		Handler: mux,
	}

	// Downloads Scanner
	scanner := fileManagement.Scanner{
		Frequency: time.Duration(cfg.Env.GetInt(env.ScanFrequency)) * time.Second,
		Directory: cfg.Env.Get(env.DownloadsPath),

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

	cfg.scanner = &scanner

	// Cull refresh tokens once a week
	cfg.refreshTokenCulling(time.Hour * 168)

	// Start server in a goroutine to allow a controlled shutdown
	go func() {
		log.Printf("Serving on: http://localhost:%s/\n", cfg.Env.Get(env.Port))
		log.Println("Starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-sigChan // Wait for shutdown

	fmt.Println()
	log.Println("Shutting down...")

	// Give the HTTP server 5 seconds to finish active requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	// FINALLY: Close the DB
	cfg.db.Close()
	log.Println("Database closed. Goodbye!")
	os.Exit(0)
}

// #region Initialization

func getFlags(args []string) (cliFlags, error) {

	flags := cliFlags{}

	for i := 0; i < len(args); i++ {

		arg := strings.ToLower(args[i])

		if arg == "--help" || arg == "-h" {
			return cliFlags{}, fmt.Errorf(helpMsg)
		}

		switch arg {
		case "-r", "--reset":
			fmt.Println("Reset flag (-r)")
			flags.dbReset = true

		case "-rb", "--reset-books":
			fmt.Println("Reset books flag (-rb)")
			flags.booksReset = true

		case "-rd", "--reset-downloads":
			fmt.Println("Reset downloads flag (-rd)")
			flags.downloadsReset = true

		case "-cl", "--clear-library":
			fmt.Println("Clear library flag (-cl)")
			flags.clearLibrary = true

		case "-cd", "--clear-downloads":
			fmt.Println("Clear downloads flag (-cd)")
			flags.clearDownloads = true

		case "-t", "--test-dataset":
			fmt.Println("Test Data Insertion flag (Resets db) (-t)")
			flags.dbReset = true
			flags.dbTestData = true

			for k := i + 1; k < len(args); k++ {
				if args[k][0] == '-' {
					i = k - 1
					break
				}

				switch args[k] {
				case "l", "test-library":
					fmt.Println("Will build test library (-t l)")
					flags.libTestData = true
				case "d", "test-downloads":
					fmt.Println("Will build test downloads (-t d)")
					flags.downTestData = true
				}
				i = k
			}

		case "-ru", "--reset-users":
			fmt.Println("Reseting users. This removes all authentication from the app! (-u)")
			flags.usersReset = true

		case "-rs", "--reset-sessions":
			fmt.Println("Reseting user sessions. (-rs)")
			flags.clearSessions = true

		default:
			return cliFlags{}, fmt.Errorf("unknown command: \"%s\". Use -h for a list of commands", arg)
		}
	}

	fmt.Println()

	return flags, nil
}

func createConfig(flags cliFlags) (*apiConfig, error) {

	// This function is so dumb it needs its own regions. Needs a refactor

	// #region Env File

	envObj, err := env.GetEnv()
	if err != nil {
		return nil, err
	}

	dataPath := envObj.Get(env.DataPath)
	metadataPath := filepath.Join(dataPath, "metadata")
	logsPath := filepath.Join(dataPath, "logs")
	dbPath := filepath.Join(dataPath, "appData.db")

	if err = os.MkdirAll(metadataPath, 0755); err != nil {
		fmt.Println("Error creating metadata folder")
		return nil, err
	}
	if err = os.MkdirAll(logsPath, 0755); err != nil {
		fmt.Println("Error creating logs folder")
		return nil, err
	}

	if err = os.MkdirAll(envObj.Get(env.DownloadsPath), 0755); err != nil {
		fmt.Println("Error creating downloads folder")
		return nil, err
	}

	if err = os.MkdirAll(envObj.Get(env.LibraryPath), 0755); err != nil {
		fmt.Println("Error creating library folder")
		return nil, err
	}

	// #region Init Flags and Database

	clearMetadata := func() error {
		if _, err := os.Stat(metadataPath); err == nil {
			files, err := os.ReadDir(metadataPath)
			if err == nil {
				for _, file := range files {
					if file.IsDir() {
						continue
					}
					err = os.Remove(filepath.Join(metadataPath, file.Name()))
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
				return err
			}
		}
		return nil
	}

	if flags.dbReset {

		// Remove the database file
		err := os.Remove(dbPath)
		if err != nil {
			return nil, err
		}

		// Clear the metadata folder
		err = clearMetadata()
		if err != nil {
			return nil, err
		}
	}

	// Initialize the database
	db, err := database.NewClient(dbPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't open database: %v", err)
	}

	if flags.clearDownloads {
		err = fileManagement.RemoveDirectoryContents(envObj.Get(env.DownloadsPath))
		if err != nil {
			return nil, err
		}
	}

	if flags.clearLibrary {
		fmt.Println("Deleting library")
		err = fileManagement.RemoveDirectoryContents(envObj.Get(env.LibraryPath))
		if err != nil {
			return nil, err
		}
	}

	// Handle the initial operations
	authRequired := false
	err = db.HandleTransaction(func(c *database.Client) error {

		// Reset users, otherwise reset sessions if the flag is enabled. Clearing users resets the sessions as well, by default
		if flags.usersReset {
			err = c.RemoveAllUsers()
			if err != nil {
				return err
			}
		} else if flags.clearSessions {
			// Clear tokens by culling using a date 61 days in the future
			err = c.CullRefreshTokens(time.Now().UTC().Add(time.Hour * 24 * 61))
			if err != nil {
				return err
			}
		}

		// Prevent redundant work
		if !flags.dbReset {
			if flags.booksReset {
				err = c.RemoveAllBooks()
				if err != nil {
					return err
				}
				err = clearMetadata()
				if err != nil {
					return err
				}
			}

			if flags.downloadsReset {
				err = c.RemoveAllDownloads()
				if err != nil {
					return err
				}
			}
		}

		// If the admin account has been created the app uses authentication
		count, err := c.CountUsers()
		if err != nil {
			return err
		}
		authRequired = count > 0

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Handle test data
	if flags.dbTestData {

		testdataPath := filepath.Join(dataPath, "test-data")
		if err = os.MkdirAll(testdataPath, 0755); err != nil {
			fmt.Println("Error creating test data folder")
			return nil, err
		}

		// Insert the rows first
		var testBooks []database.Book
		err = db.HandleTransaction(func(c *database.Client) error {
			testBooks, err = c.InsertTestData(metadataPath, testdataPath)
			if err != nil {
				return err
			}
			return nil
		})

		// Create the test Library
		testLibrarySize := 30
		if flags.libTestData {

			// Inserts 30 books with random text and audio files. All books will have a cover.
			// Half will have a metadata file.

			fmt.Println()
			fmt.Println("======= Creating test library =======")

			handleDirs := func(book database.Book, useAuthor, useSeries bool) (string, error) {

				author := ""
				if useAuthor {
					author = "Unknown"
					if len(book.Authors) > 0 {
						author = book.Authors[0].Name
					}
				}

				series := ""
				title := book.Title
				if useSeries && len(book.Series) > 0 {
					series = book.Series[0].Name
					// Randomly add the series index
					if book.Series[0].Index != nil && rand.Float32() > 0.5 {
						title = *book.Series[0].Index + " - " + title
					}
				}

				bookPath := filepath.Join(author, series, title)

				return bookPath, nil
			}

			offset := testLibrarySize / 2

			// No metadata
			for i := 0; i < offset; i++ {

				bookPath, err := handleDirs(testBooks[i], rand.Float32() > 0.5, rand.Float32() > 0.5)
				if err != nil {
					return nil, err
				}

				testBooks[i].Files, err = fileManagement.CreateTestDirectory(
					bookPath,
					envObj.Get(env.LibraryPath),
					nil,
					filepath.Join(metadataPath, testBooks[i].Id.String()+".jpg"), rand.Intn(10)+1, rand.Intn(10)+1,
				)
				if err != nil {
					testBooks[i].Files = fileManagement.Files{}
					fmt.Println("Failed to build test folder for \"" + bookPath + "\": " + err.Error())
				}
			}

			// With metadata
			for i := offset; i < testLibrarySize; i++ {

				bookPath, err := handleDirs(testBooks[i], rand.Float32() > 0.5, rand.Float32() > 0.5)
				if err != nil {
					return nil, err
				}

				testBooks[i].Files, err = fileManagement.CreateTestDirectory(
					bookPath,
					envObj.Get(env.LibraryPath),
					metadata.BookToMetadata(testBooks[i]),
					filepath.Join(metadataPath, testBooks[i].Id.String()+".jpg"), rand.Intn(10)+1, rand.Intn(10)+1,
				)
				if err != nil {
					testBooks[i].Files = fileManagement.Files{}
					fmt.Println("Failed to build test folder for \"" + bookPath + "\": " + err.Error())
				}
			}
		}

		// Create test downloads
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
					d.cover = filepath.Join(metadataPath, d.cover)
				}

				_, err = fileManagement.CreateTestDirectory(
					filepath.Join(d.title),
					envObj.Get(env.DownloadsPath),
					d.md,
					d.cover,
					d.audio, d.text,
				)
				if err != nil {
					return nil, err
				}
			}
		}

		err = db.HandleTransaction(func(c *database.Client) error {

			for _, book := range testBooks[:testLibrarySize] {
				err = book.ApplyBookFiles(c)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// #region Return apiConfig

	return &apiConfig{
		db:      db,
		mdCache: cache.NewCache(time.Minute * 5),

		Env: envObj,

		metadataPath: metadataPath,
		logsPath:     logsPath,

		libraryName:   "/media/library",
		downloadsName: "/media/downloads",

		authRequired: authRequired,
	}, nil
}

// #region Routines

func (cfg *apiConfig) refreshTokenCulling(frequency time.Duration) {
	go func(ctx context.Context) {
		for {
			log.Println("Removing expired refresh tokens...")

			select {
			case <-ctx.Done():
				return

			default:
				err := cfg.db.HandleTransaction(func(c *database.Client) error { return c.CullRefreshTokens(time.Now().UTC()) })
				if err != nil {
					log.Println(err)
				}
			}

			time.Sleep(frequency)
		}
	}(context.Background())
}

// #region Utility

func authenticate(required bool, r *http.Request, secret string) (uuid.UUID, error) {

	if !required {
		return uuid.Nil, nil
	}

	token, err := auth.GetBearerToken(r)
	if err != nil {
		return uuid.Nil, err
	}

	id, err := auth.ValidateJWT(token, secret)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func handlerFrontend(w http.ResponseWriter, r *http.Request) {

	fHandler := http.FileServer(http.Dir("./static/"))
	fullPath := filepath.Join("./static", filepath.Clean(r.URL.Path))

	// Check if the requested file exists
	info, err := os.Stat(fullPath)

	// If the asset exists on disk and is a file, serve it directly (CSS, JS, PNGs)
	if err == nil && !info.IsDir() {
		fHandler.ServeHTTP(w, r)
		return
	}

	// SPA FALLBACK: The file doesn't exist, meaning it's a Vue client path (e.g., /login).
	// Serve index.html and let Vue Router handle the user destination.
	http.ServeFile(w, r, "./static/index.html")
}

func (cfg *apiConfig) handlerHealthCheckfunc(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	err := cfg.db.PingContext(ctx)
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "UNHEALTHY: Database connection failed", nil)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// #region Middleware

func (cfg *apiConfig) uuidMiddleware(handler func(uuid.UUID, http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		idStr := r.PathValue("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid id", err)
			return
		}

		_, err = authenticate(cfg.authRequired, r, cfg.Env.Get(env.TokenSecret))
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, AuthBadAuthorization, err)
			return
		}

		handler(id, w, r)
	}

}

func (cfg *apiConfig) authMiddleware(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		_, err := authenticate(cfg.authRequired, r, cfg.Env.Get(env.TokenSecret))
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, AuthBadAuthorization, err)
			return
		}

		handler(w, r)
	}
}

// #region Help Message
// It's down here because it kept screwing up the collapsed functions and making editing this file annoying

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
	This also deletes all users, and disables authentication.

[-rb, --reset-books]
	Clears the books table, which will also clear all the category tables.

[-rd, --reset-downloads]
	Clears the downloads table

[-ru, --reset-users]:
	Clears the users table, which also removes any authentication in the app. Use with caution.

[-rs, --reset-sessions]
	Clears the refresh tokens, effectively logging out every user with an expired session.

[-cl, --clear-library]:
	Clears the contents of the library, defined in the .env as LIBRARY_PATH.

[-cd, --clear-downloads]:
	Clears the contents of the downloads, defined in the .env as DOWNLOAD_PATH.

[-t, --test-dataset] <l, test-library> <d, test-downloads>:
	Does everything reset does and inserts the test dataset (./backend/internal/database/testData.go).

	<l, test-library>:
		Clears the library and creates a test library structure using the test dataset.
		Inserts 30 books with random text and audio files. All books will have a cover.
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
