package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/Ethanol2/book-organizer/internal/database"
	"github.com/Ethanol2/book-organizer/internal/fileManagement"
	"github.com/Ethanol2/book-organizer/internal/metadata"
	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
)

func (cfg *apiConfig) handlerGetBook(id uuid.UUID, w http.ResponseWriter, r *http.Request) {

	book, err := cfg.db.GetBook(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error", err)
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
		results, err := cfg.db.GetBooks(r.URL.Query())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Database error", err)
			return
		}

		for i := range results.Items {
			results.Items[i].Files.Prepend(cfg.libraryName)
		}

		log.Println("Fetching book details")

		respondWithJson(w, http.StatusOK, results)

	case "":
	case "summary":
		results, err := cfg.db.GetBooksSummary(r.URL.Query())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Database error", err)
			return
		}

		for i := range results.Items {
			if results.Items[i].Cover != nil {
				cover := *results.Items[i].Cover
				cover = path.Join(cfg.libraryName, cover)
				results.Items[i].Cover = &cover
			}
		}

		log.Println("Fetching book summaries")

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
		respondWithError(w, http.StatusBadRequest, "Couldn't read body", err)
		return
	}

	var coverFile *os.File
	if params.Cover != nil {
		coverFile, err = fileManagement.DownloadTempFile(*params.Cover)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Failed to fetch cover from url. Only png and jpg are currently supported", err)
			log.Println("Image:", *params.Cover)
			return
		}
		defer coverFile.Close()
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

	book, err := cfg.db.AddBook(params)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				respondWithError(w, http.StatusBadRequest, "Books can't share ISBN or ASIN numbers to prevent duplicates", err)
				return
			}
		}

		respondWithError(w, http.StatusInternalServerError, "Couldn't add book to db", err)
		return
	}

	if coverFile != nil {
		err = fileManagement.MoveFilesWithPaths(coverFile.Name(), path.Join(cfg.metadataPath, book.Id.String()+path.Ext(coverFile.Name())))
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "failed to create cover file", err)
			return
		}
	}

	respondWithJson(w, http.StatusOK, book)
}

func (cfg *apiConfig) handlerUpdateBook(id uuid.UUID, w http.ResponseWriter, r *http.Request) {

	var params database.BookParams
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to read body", err)
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

	// Begining db transaction here in case the file moving doesn't work
	err = cfg.db.Begin()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database start transaction error", err)
		return
	}
	defer cfg.db.Rollback()

	oldPath, err := cfg.db.GetBookDirectory(id)

	book, needsFileUpdate, err := cfg.db.UpdateBook(id, params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update database", err)
		return
	}

	if needsFileUpdate {
		authorDir, seriesDir, bookDir, err := cfg.db.GetPathComponents(id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error moving files", err)
			return
		}

		err = fileManagement.CreateDirectory(path.Join(cfg.libraryPath, authorDir))
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error moving files", err)
			return
		}

		err = fileManagement.CreateDirectory(path.Join(cfg.libraryPath, authorDir, seriesDir))
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error moving files", err)
			return
		}

		newPath := path.Join(authorDir, seriesDir, bookDir)

		if book.Files.Root != &newPath {
			err = fileManagement.MoveFilesWithPaths(path.Join(cfg.libraryPath, *oldPath), path.Join(cfg.libraryPath, newPath))
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Error moving files", err)
				return
			}
		}
	}

	err = cfg.db.Commit()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database commit transaction error", err)
		return
	}

	book, err = cfg.db.GetBook(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve the updated book from the database", err)
		return
	}

	if book.Files.Root != nil {
		fileManagement.CreateMetadataFile(*metadata.BookToMetadata(book), path.Join(cfg.libraryPath, *book.Files.Root))
	}

	book.Files.Prepend(cfg.libraryName)
	respondWithJson(w, http.StatusOK, book)
}

func (cfg *apiConfig) handlerUpdateBookCover(id uuid.UUID, w http.ResponseWriter, r *http.Request) {

	tmp, err := fileManagement.CreateTempFileFromRequest(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to download image", err)
		return
	}
	defer tmp.Close()

	oldPath, newPath, err := cfg.db.UpdateBookCover(id, path.Ext(tmp.Name()))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error", err)
		return
	}

	if oldPath == "" {
		oldPath = path.Join(cfg.metadataPath, id.String()+path.Ext(tmp.Name()))
	}

	err = os.Remove(oldPath)
	if err != nil {
		log.Println("Failed to remove old cover")
		respondWithError(w, http.StatusInternalServerError, "File error", err)
		return
	}

	if newPath == "" {
		newPath = path.Join(cfg.metadataPath, id.String()+path.Ext(tmp.Name()))
	}

	err = fileManagement.MoveFilesWithPaths(tmp.Name(), newPath)
	if err != nil {
		log.Println("Failed to move new cover to path")
		respondWithError(w, http.StatusInternalServerError, "File error", err)
		return
	}

	book, err := cfg.db.GetBook(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error", err)
		return
	}

	respondWithJson(w, http.StatusOK, book)

}

func (cfg *apiConfig) handlerGetBookCover(id uuid.UUID, w http.ResponseWriter, r *http.Request) {

	book, err := cfg.db.GetBook(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error", err)
		return
	}
	if book.Id == nil {
		respondWithError(w, http.StatusBadRequest, "Book not found", err)
		return
	}
	if book.Files.Cover == nil {
		respondWithError(w, http.StatusNoContent, "Book doesn't have a cover associated", err)
		return
	}

	coverPath := ""
	if book.Files.Root == nil {
		coverPath = path.Join(cfg.metadataPath, *book.Files.Cover)
	} else {
		author, series, _, err := cfg.db.GetPathComponents(id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Database error", err)
			return
		}

		coverPath = path.Join(cfg.libraryPath, author, series, *book.Files.Cover)
	}
	log.Println("Serving book cover from", coverPath)

	http.ServeFile(w, r, coverPath)
}

func (cfg *apiConfig) handlerDeleteBook(id uuid.UUID, w http.ResponseWriter, r *http.Request) {

	if exists, err := cfg.db.CheckBookExistsID(id); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong while querying the database", err)
		return
	} else if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if deleteFiles := r.URL.Query().Get("files"); deleteFiles == "true" {
		log.Println("Deleting files for \"", id, "\"")
		dir, err := cfg.db.GetBookDirectory(id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't delete book files", err)
			return
		}
		if dir != nil {
			err = fileManagement.DeleteFiles(path.Join(cfg.libraryPath, *dir))
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Couldn't delete book files", err)
				return
			}
		}
		err = cfg.db.UpdateBookFiles(id, fileManagement.Files{})
	}

	if deleteBook := r.URL.Query().Get("book"); deleteBook == "true" {

		// Delete metedata cover
		err := fileManagement.DeleteFiles(path.Join(cfg.metadataPath, id.String()+".jpg"))
		if err != nil {
			log.Println(err)
		}

		err = cfg.db.DeleteBook(id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to delete book from the database", err)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerGetScanLibrary(w http.ResponseWriter, r *http.Request) {

	libraryParams := map[database.BookParams]fileManagement.Files{}

	// Get existing library to prevent duplicates
	_, dirs, err := cfg.db.GetAllBooksDirectories()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong while preparing for the scan", err)
		return
	}
	// Get base since the roots include author and series
	for i := range dirs {
		dirs[i] = path.Base(dirs[i])
	}

	scanErrors := []string{}
	var folderScan func(string, string, []string)
	folderScan = func(dirPrefix string, dir string, names []string) {

		currentPath := path.Join(dirPrefix, dir)
		if dir != "" {
			names = append(names, dir)
		}

		// Check the recursion hasn't gone too deep -> Author/Series/Book -> max 3 folder levels
		if len(names) > 3 {
			err := fmt.Sprintf("scan max depth (3) exceeded: %d", len(names))
			log.Println(err)
			scanErrors = append(scanErrors, err)
			return
		}

		// List for the current folder
		dirItems := []fileManagement.Files{}
		adder := func(files []fileManagement.Files) error {
			dirItems = append(dirItems, files...)
			return nil
		}

		// Get the contents of the current folder
		scanner := fileManagement.Scanner{
			Directory:  currentPath,
			AddHandler: adder,
		}
		err := scanner.ScanNew(dirs)
		if err != nil {
			strErr := fmt.Sprintln("Error trying to read the folder at \"", currentPath, "\" =>", err)
			log.Print(strErr)
			scanErrors = append(scanErrors, strErr)
			return
		}

		// For each item in the folder
		for _, item := range dirItems {

			// If the item has no audio or text files then scan nested folders, since we aren't in a book
			if item.HasNoFiles() {
				// Check if the folder has sub folders
				if item.Directories != nil {
					folderScan(currentPath, *item.Root, names)
				}
				continue
			}

			if item.Root == nil {
				err := fmt.Sprintln("An item with no root appeared in \"", dirPrefix, "\"")
				log.Print(err)
				scanErrors = append(scanErrors, err)
				continue
			}

			p := path.Join(names...)

			root := path.Join(p, *item.Root)
			item.Root = &root

			if item.Cover != nil {
				c := path.Join(p, *item.Cover)
				item.Cover = &c
			}

			// If the item has a metadata file then import that and continue
			if item.HasMetadata {
				file, err := os.Open(path.Join(cfg.libraryPath, *item.Root, "metadata.json"))
				if err != nil {
					strErr := fmt.Sprintln("Error trying to open metadata file in \"", *item.Root, "\" =>", err)
					log.Print(strErr)
					scanErrors = append(scanErrors, strErr)
					continue
				}

				var md fileManagement.MetadataFile
				err = json.NewDecoder(file).Decode(&md)
				file.Close()
				if err != nil {
					strErr := fmt.Sprintln("Error trying to decode metadata file in \"", *item.Root, "\" =>", err)
					log.Print(strErr)
					scanErrors = append(scanErrors, strErr)
					continue
				}

				libraryParams[metadata.MetadataToBookParams(md)] = item
				continue
			}

			title := path.Base(*item.Root)

			var authorCat []database.Category
			authorCat = []database.Category{{Name: names[0]}}

			var seriesCat []database.Category
			if len(names) > 1 {

				var index *string
				split := strings.SplitN(title, " - ", 2)
				if len(split) > 1 {
					index = &split[0]
					title = split[1]
				}

				seriesCat = []database.Category{{Name: names[1], Index: index}}
			}

			params := database.BookParams{
				Title:   &title,
				Authors: &authorCat,
				Series:  &seriesCat,
			}

			libraryParams[params] = item
		}
	}

	log.Println("Starting library scan")
	folderScan(cfg.libraryPath, "", []string{})

	applyBookFiles := func(book database.Book, files fileManagement.Files) error {
		log.Println("Inserting book files")
		err = cfg.db.Begin()
		if err != nil {
			return err
		}
		defer cfg.db.Rollback()

		book.Files = files
		err = book.ApplyBookFiles(cfg.db)
		if err != nil {
			return err
		}

		err = cfg.db.Commit()
		if err != nil {
			return err
		}
		return nil
	}
	updateExistingBook := func(id uuid.UUID, params database.BookParams, files fileManagement.Files) error {
		hasFiles, err := cfg.db.CheckBookHasFiles(id)
		if err != nil {
			scanErrors = append(scanErrors, "Something went wrong checking for files associated with ", *params.Title)
			return nil
		}

		if hasFiles {
			scanErrors = append(scanErrors, fmt.Sprintf("Duplicate files for %s exist at %s", *params.Title, *files.Root))
			return nil
		}

		book, _, err := cfg.db.UpdateBook(id, params)
		if err != nil {
			return err
		}

		err = applyBookFiles(book, files)
		if err != nil {
			return err
		}
		return nil
	}

	for params, files := range libraryParams {

		if params.ISBN != nil {
			if strings.TrimSpace(*params.ISBN) == "" {
				params.ISBN = nil
			} else if !metadata.IsValidISBN13(*params.ISBN) {
				scanErrors = append(scanErrors, "ISBN not valid => "+*files.Root)
				continue
			} else if ok, id, _ := cfg.db.CheckBookExistsISBN(*params.ISBN); ok {

				err = updateExistingBook(id, params, files)
				if err != nil {
					respondWithError(w, http.StatusInternalServerError, "Error adding books to the database", err)
					return
				}
				continue
			}
		}

		if params.ASIN != nil {
			if strings.TrimSpace(*params.ASIN) == "" {
				params.ASIN = nil
			} else if !metadata.IsValidASIN(*params.ASIN) {
				scanErrors = append(scanErrors, "ASIN not valid => "+*files.Root)
				continue
			} else if ok, id, _ := cfg.db.CheckBookExistsASIN(*params.ISBN); ok {

				err = updateExistingBook(id, params, files)
				if err != nil {
					respondWithError(w, http.StatusInternalServerError, "Error adding books to the database", err)
					return
				}
				continue
			}
		}

		log.Println("Adding \"", *params.Title, "\" to the database")
		book, err := cfg.db.AddBook(params)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error adding books to the database", err)
			return
		}

		err = applyBookFiles(book, files)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error adding books to the database", err)
			return
		}
	}

	log.Println("Library Scan Complete")

	results, err := cfg.db.GetBooksSummary(r.URL.Query())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving book summaries after scan", err)
		return
	}

	for i := range results.Items {
		if results.Items[i].Cover != nil {
			cover := *results.Items[i].Cover
			cover = path.Join(cfg.libraryName, cover)
			results.Items[i].Cover = &cover
		}
	}

	respondWithJson(w, http.StatusOK, struct {
		Results database.BookSearchResults[[]database.BookOverview] `json:"results"`
		Errors  []string                                            `json:"errors"`
	}{Results: results, Errors: scanErrors})
}
