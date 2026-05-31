package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/Ethanol2/book-organizer/internal/database"
	"github.com/Ethanol2/book-organizer/internal/env"
	"github.com/Ethanol2/book-organizer/internal/fileManagement"
	"github.com/Ethanol2/book-organizer/internal/metadata"
	"github.com/google/uuid"
	"github.com/ncruces/go-sqlite3"
)

func (cfg *apiConfig) handlerGetBook(id uuid.UUID, w http.ResponseWriter, r *http.Request) {

	book, err := cfg.db.GetBook(id)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithJson(w, http.StatusNotFound, struct{}{})
			return
		}
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	if book.Id == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	book.Files.Prepend(cfg.libraryName)

	log.Println("Fetching \"", book.Title, "\" book details")

	respondWithJson(w, http.StatusOK, book)
}

func (cfg *apiConfig) handlerGetBooks(w http.ResponseWriter, r *http.Request) {

	getFullResults := r.URL.Query().Get("view")

	switch getFullResults {
	case "full":
		log.Println("Fetching book details")

		results, err := cfg.db.GetBooks(r.URL.Query())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
			return
		}

		for i := range results.Items {
			results.Items[i].Files.Prepend(cfg.libraryName)
		}

		respondWithJson(w, http.StatusOK, results)

	case "":
	case "summary":
		log.Println("Fetching book summaries")

		results, err := cfg.db.GetBooksSummary(r.URL.Query())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
			return
		}

		for i := range results.Items {
			if results.Items[i].Cover != nil {
				cover := *results.Items[i].Cover
				cover = path.Join(cfg.libraryName, cover)
				results.Items[i].Cover = &cover
			}
		}

		respondWithJson(w, http.StatusOK, results)

	default:
		respondWithError(
			w,
			http.StatusBadRequest,
			"View options are 'full' for full details or 'summary' for just title, subtitle, cover and authors",
			fmt.Errorf("Invalid view value when fetching books: %s", getFullResults))
	}
}

func (cfg *apiConfig) handlerPostBook(w http.ResponseWriter, r *http.Request) {

	var params database.BookParams
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, BodyDecodeError, err)
		return
	}

	var coverFile *os.File
	if params.Cover != nil {
		if !strings.HasPrefix(*params.Cover, cfg.downloadsName) && !strings.HasPrefix(*params.Cover, cfg.libraryName) {

			coverFile, err = fileManagement.DownloadTempFile(*params.Cover)
			if err != nil {
				respondWithError(w, http.StatusBadRequest, CoverURLError, err)
				log.Println("Image:", *params.Cover)
				return
			}
			defer coverFile.Close()
		}
	}

	if params.ISBN != nil {
		if *params.ISBN == "" {
			params.ISBN = nil
		} else if !metadata.IsValidISBN13(*params.ISBN) {
			respondWithError(w, http.StatusBadRequest, "Invalid ISBN", nil)
			return
		}
	}
	if params.ASIN != nil {
		if *params.ASIN == "" {
			params.ASIN = nil
		} else if !metadata.IsValidASIN(*params.ASIN) {
			respondWithError(w, http.StatusBadRequest, "Invalid ASIN", nil)
			return
		}
	}

	var book database.Book
	err = cfg.db.HandleTransaction(func(c *database.Client) error {
		book, err = c.AddBook(params)
		return err
	})
	if err != nil {
		var sqliteErr *sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code() == sqlite3.CONSTRAINT {
				respondWithError(w, http.StatusBadRequest, "Books can't share ISBN or ASIN numbers to prevent duplicates", err)
				return
			}
		}
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	if coverFile != nil {
		err = fileManagement.MoveFiles(coverFile.Name(), path.Join(cfg.metadataPath, book.Id.String()+path.Ext(coverFile.Name())))
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, FileMoveError, err)
			return
		}
	}

	respondWithJson(w, http.StatusOK, book)
}

func (cfg *apiConfig) handlerUpdateBook(id uuid.UUID, w http.ResponseWriter, r *http.Request) {

	var params struct {
		database.BookParams
		MoveFiles bool `json:"move_files"`
	}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, BodyDecodeError, err)
		return
	}

	if params.ISBN != nil {
		if *params.ISBN == "" {
			params.ISBN = nil
		} else if !metadata.IsValidISBN13(*params.ISBN) {
			respondWithError(w, http.StatusBadRequest, "Invalid ISBN", nil)
			return
		}
	}
	if params.ASIN != nil {
		if *params.ASIN == "" {
			params.ASIN = nil
		} else if !metadata.IsValidASIN(*params.ASIN) {
			respondWithError(w, http.StatusBadRequest, "Invalid ASIN", nil)
			return
		}
	}

	var newCover *os.File
	if params.Cover != nil {
		newCover, err = fileManagement.DownloadTempFile(*params.Cover)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, CoverURLError, err)
			log.Println("Image:", *params.Cover)
			return
		}
		defer newCover.Close()
	}

	oldPath, err := cfg.db.GetBookDirectory(id)

	var book database.Book
	needsFileUpdate := false

	err = cfg.db.HandleTransaction(func(c *database.Client) error {
		book, needsFileUpdate, err = c.UpdateBook(id, params.BookParams)
		return err
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	if needsFileUpdate && params.MoveFiles {
		newPath, err := cfg.db.GetBookPath(id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, FileMoveError, err)
			return
		}

		if book.Files.Root != &newPath {
			err = fileManagement.MoveFiles(path.Join(cfg.Env.Get(env.LibraryPath), *oldPath), path.Join(cfg.Env.Get(env.LibraryPath), newPath))
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, FileMoveError, err)
				return
			}
		}
	}

	if newCover != nil {

		coverPath := ""

		if book.Files.Root == nil {
			coverPath = path.Join(cfg.metadataPath, book.Id.String()+path.Ext(newCover.Name()))
		} else {
			coverPath = path.Join(cfg.Env.Get(env.LibraryPath), *book.Files.Root, "cover.jpg")
		}

		err = fileManagement.DeleteFiles(coverPath)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, FileDeleteError, err)
			return
		}

		err = fileManagement.MoveFiles(newCover.Name(), coverPath)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, FileMoveError, err)
		}
	}

	book, err = cfg.db.GetBook(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	if book.Files.Root != nil {
		fileManagement.CreateMetadataFile(*metadata.BookToMetadata(book), path.Join(cfg.Env.Get(env.LibraryPath), *book.Files.Root))
	}

	book.Files.Prepend(cfg.libraryName)
	respondWithJson(w, http.StatusOK, book)
}

func (cfg *apiConfig) handlerDeleteBook(id uuid.UUID, w http.ResponseWriter, r *http.Request) {

	if exists, err := cfg.db.CheckBookExistsID(id); err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	} else if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if deleteFiles := r.URL.Query().Get("files"); deleteFiles == "true" {
		log.Println("Deleting files for \"", id, "\"")
		dir, err := cfg.db.GetBookDirectory(id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, FileDeleteError, err)
			return
		}
		if dir != nil {
			err = fileManagement.DeleteFiles(path.Join(cfg.Env.Get(env.LibraryPath), *dir))
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, FileDeleteError, err)
				return
			}
		}
		err = cfg.db.HandleTransaction(func(c *database.Client) error {
			return c.UpdateBookFiles(id, fileManagement.Files{})
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		}
	}

	if deleteBook := r.URL.Query().Get("book"); deleteBook == "true" {

		// Delete metedata cover
		err := fileManagement.DeleteFiles(path.Join(cfg.metadataPath, id.String()+".jpg"))
		if err != nil {
			log.Println(err)
		}

		err = cfg.db.HandleTransaction(func(c *database.Client) error {
			return c.DeleteBook(id)
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerGetScanLibrary(w http.ResponseWriter, r *http.Request) {

	// This function is a monster, hence the more detailed commenting

	// Types ==========================================================================
	type foldersIds struct {
		Folders []string
		Ids     []uuid.UUID
	}
	type scanError struct {
		Type  string `json:"type"`
		Path  string `json:"path"`
		Error string `json:"error"`
	}

	// Global vars
	libraryParams := map[database.BookParams]fileManagement.Files{} // New Books in the library
	knownDirs := map[string]foldersIds{}                            // Known directories in the library, mapped to their parent folders
	scanErrors := []scanError{}                                     // Errors that occur during the scan that aren't cause for panic. The list is returned in the response

	// Start the transaction
	err := cfg.db.HandleTransaction(func(c *database.Client) error {

		// Functions ======================================================================

		// The function used by the scanner to address deleted directories
		deleteHandler := func(id uuid.UUID) error {
			return c.UpdateBookFiles(id, fileManagement.Files{})
		}

		// Used by the scanner to post file changes. Adds the full path to the files before being forwarded to the actual update
		updateHandler := func(prefix string) func(id uuid.UUID, files fileManagement.Files) error {
			return func(id uuid.UUID, files fileManagement.Files) error {
				files.Prepend(prefix)
				return c.UpdateBookFiles(id, files)
			}
		}

		// The main scanning logic. The function is declared and defined seperately to allow recursion.
		var folderScan func(string, string, []string)
		folderScan = func(pathPrefix string, currentDirectory string, pathComponents []string) {

			// Check the recursion hasn't gone too deep -> Author/Series/Book -> max 3 folder levels
			if len(pathComponents) > 3 {
				err := fmt.Sprintf("scan max depth (3) exceeded: %d", len(pathComponents))
				log.Println(err)
				scanErrors = append(scanErrors,
					scanError{
						Type:  "Processing",
						Path:  path.Join(pathPrefix, currentDirectory),
						Error: "maximum recursion depth exceeded",
					})
				return
			}

			// Create the current path using the prefix and the current directory.
			// If the current directory isn't empty, add it to the list of path components
			currentPath := path.Join(pathPrefix, currentDirectory)
			if currentDirectory != "" {
				pathComponents = append(pathComponents, currentDirectory)
			}

			// List for the current folder
			dirItems := []fileManagement.Files{}

			// handler functions
			addHandler := func(files []fileManagement.Files) error {
				dirItems = append(dirItems, files...)
				return nil
			}

			relativePath := path.Join(pathComponents...)

			// Initialize the scanner
			scanner := fileManagement.Scanner{
				Directory:     currentPath,
				AddHandler:    addHandler,
				UpdateHandler: updateHandler(relativePath),
				DeleteHandler: deleteHandler,

				IncludeEmpty: true,
			}

			// Scan new folders. Use the known folders as the ignore list
			err := scanner.ScanNew(knownDirs[currentPath].Folders)
			if err != nil {
				strErr := fmt.Sprintln("Error trying to read the folder at \"", currentPath, "\" =>", err)
				log.Print(strErr)
				scanErrors = append(scanErrors, scanError{
					Type:  "Read Error",
					Path:  currentPath,
					Error: err.Error(),
				})
				return
			}

			// Scan known folders
			err = scanner.ScanExisting(knownDirs[currentPath].Ids, knownDirs[currentPath].Folders)
			if err != nil {
				strErr := fmt.Sprintln("Error trying to update books at \"", currentPath, "\" =>", err)
				log.Print(strErr)
				scanErrors = append(scanErrors, scanError{
					Type:  "Read Error",
					Path:  currentPath,
					Error: err.Error(),
				})
				return
			}

			// For each item in the folder
			for _, item := range dirItems {

				// If the item has no audio or text files then scan nested folders, since we aren't in a book
				if item.HasNoFiles() {
					// Check if the folder has sub folders
					if item.Directories != nil {
						folderScan(currentPath, *item.Root, pathComponents)
					}
					continue
				}

				// This shouldn't happen
				if item.Root == nil {
					err := fmt.Sprintln("An item with no root appeared in \"", pathPrefix, "\"")
					log.Print(err)
					scanErrors = append(scanErrors, scanError{
						Type:  "Invalid Data",
						Path:  currentPath,
						Error: "item found with no root",
					})
					continue
				}

				// Get the path to this item
				root := path.Join(relativePath, *item.Root)
				item.Root = &root

				// If there's a cover present, get the path for that
				if item.Cover != nil {
					c := path.Join(relativePath, *item.Cover)
					item.Cover = &c
				}

				// If the item has a metadata file then import that and continue
				if item.HasMetadata {
					md, err := fileManagement.OpenMetadataFile(path.Join(cfg.Env.Get(env.LibraryPath), *item.Root, "metadata.json"))
					if err != nil {
						strErr := fmt.Sprintln("Error trying to open metadata file in \"", *item.Root, "\" =>", err)
						log.Print(strErr)
						scanErrors = append(scanErrors, scanError{
							Type:  "Read Error",
							Path:  *item.Root,
							Error: err.Error(),
						})
						continue
					}

					libraryParams[metadata.MetadataToBookParams(*md)] = item
					continue
				}

				// Get the book title from the folder
				title := path.Base(*item.Root)

				// Get the author from the topmost folder
				authors := []database.Category{}
				if len(pathComponents) > 0 {
					authors = append(authors, database.Category{Name: pathComponents[0]})
				}

				// Get the series from the second folder. Extract the series index from the title, if it exists.
				var series []database.Category
				if len(pathComponents) > 1 {

					var index *string
					split := strings.SplitN(title, " - ", 2)
					if len(split) > 1 {
						index = &split[0]
						title = split[1]
					}

					series = []database.Category{{Name: pathComponents[1], Index: index}}
				}

				libraryParams[database.BookParams{
					Title:   &title,
					Authors: &authors,
					Series:  &series,
				}] = item
			}
		}

		// Used when a book is matched to an untracked folder
		updateExistingBook := func(id uuid.UUID, params database.BookParams, files fileManagement.Files) error {
			hasFiles, err := c.CheckBookHasFiles(id)
			if err != nil {
				return fmt.Errorf("Something went wrong checking for files associated with %s", *params.Title)
			}

			if hasFiles {
				return fmt.Errorf("Duplicate files for %s exist at %s", *params.Title, *files.Root)
			}

			book, _, err := c.UpdateBook(id, params)
			if err != nil {
				return err
			}

			book.Files = files
			err = book.ApplyBookFiles(c)
			if err != nil {
				return err
			}

			return nil
		}

		// Main Logic ======================================================================

		fmt.Println()
		log.Println("Starting library scan")

		// Get existing library to prevent duplicates
		if ids, knownDirsList, err := c.GetAllBooksDirectories(); err != nil {
			return err
		} else {
			// Format known directories to be handled by the scan function
			for i := range knownDirsList {
				dir, name := path.Split(knownDirsList[i])
				dir = path.Join(cfg.Env.Get(env.LibraryPath), dir)
				if item, ok := knownDirs[dir]; ok {
					item.Folders = append(item.Folders, name)
					item.Ids = append(item.Ids, ids[i])
					knownDirs[dir] = item
				} else {
					knownDirs[dir] = foldersIds{
						Folders: []string{name},
						Ids:     []uuid.UUID{ids[i]},
					}
				}
			}
		}

		folderScan(cfg.Env.Get(env.LibraryPath), "", []string{})

		// Handle new books
		for params, files := range libraryParams {

			// If the book has metadata and an ISBN 13 number, validate it and try to match it with an existing book
			if params.ISBN != nil {

				if *params.ISBN == "" {
					params.ISBN = nil
				} else {

					if !metadata.IsValidISBN13(*params.ISBN) {
						scanErrors = append(scanErrors, scanError{
							Type:  "Invalid Data",
							Path:  *files.Root,
							Error: "invalid ISBN-13 number: " + *params.ISBN,
						})
						continue
					} else if ok, id, _ := c.CheckBookExistsISBN(*params.ISBN); ok {

						err := updateExistingBook(id, params, files)
						if err != nil {
							scanErrors = append(scanErrors, scanError{
								Type:  "Update Error",
								Path:  *files.Root,
								Error: err.Error(),
							})
						}
						continue
					}
				}
			}

			// Do the same with the ASIN number
			if params.ASIN != nil {

				if *params.ASIN == "" {
					params.ASIN = nil
				} else {

					if !metadata.IsValidASIN(*params.ASIN) {
						scanErrors = append(scanErrors, scanError{
							Type:  "Invalid Data",
							Path:  *files.Root,
							Error: "invalid ASIN number: " + *params.ASIN,
						})
						continue
					} else if ok, id, _ := c.CheckBookExistsASIN(*params.ASIN); ok {

						err := updateExistingBook(id, params, files)
						if err != nil {
							scanErrors = append(scanErrors, scanError{
								Type:  "Update Error",
								Path:  *files.Root,
								Error: err.Error(),
							})
						}
						continue
					}
				}
			}

			// Add the new book to the database
			log.Println("Adding \"", *params.Title, "\" to the database")
			book, err := c.AddBook(params)
			if err != nil {
				return err
			}

			// Add the book files to the book
			book.Files = files
			err = book.ApplyBookFiles(c)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	log.Println("Library Scan Complete")

	// Return the book summaries using the search params
	results, err := cfg.db.GetBooksSummary(r.URL.Query())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	// Prepend the library endpoint path to the cover paths
	for i := range results.Items {
		if results.Items[i].Cover != nil {
			cover := *results.Items[i].Cover
			cover = path.Join(cfg.libraryName, cover)
			results.Items[i].Cover = &cover
		}
	}

	// Send response
	respondWithJson(w, http.StatusOK, struct {
		database.BookSearchResults[[]database.BookOverview] `json:"results"`
		Errors                                              []scanError `json:"errors"`
	}{BookSearchResults: results, Errors: scanErrors})
}
