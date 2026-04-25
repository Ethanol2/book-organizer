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

	"github.com/Ethanol2/book-organizer/internal/fileManagement"
	"github.com/Ethanol2/book-organizer/internal/metadata"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerGetDownloads(w http.ResponseWriter, r *http.Request) {

	log.Println("Fetching downloads")
	downloads, err := cfg.db.GetDownloads()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "something went wrong retrieving downloads", err)
	}

	for i := range downloads {
		downloads[i].Files.Prepend(cfg.downloadsName)
	}

	respondWithJson(w, http.StatusOK, downloads)
}

func (cfg *apiConfig) handlerGetDownload(id uuid.UUID, w http.ResponseWriter, r *http.Request) {

	download, err := cfg.db.GetDownload(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get download", err)
		return
	}

	download.Files.Prepend(cfg.downloadsName)

	respondWithJson(w, http.StatusOK, download)
}

func (cfg *apiConfig) handlerAssociateDownloadToBook(downloadId uuid.UUID, w http.ResponseWriter, r *http.Request) {

	downloadDir, err := cfg.db.GetDownloadDir(downloadId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusBadRequest, "Download not found", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Database error", err)
		return
	}

	var bookIdStruct struct {
		BookId             uuid.UUID `json:"book_id"`
		UseDownloadedCover bool      `json:"use_downloaded_cover"`
	}
	err = json.NewDecoder(r.Body).Decode(&bookIdStruct)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't read body", err)
		return
	}

	bookExists, err := cfg.db.CheckBookExists(bookIdStruct.BookId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error", err)
		return
	}
	if !bookExists {
		respondWithError(w, http.StatusBadRequest, "Book not found", err)
		return
	}

	authorDir, seriesDir, bookDir, err := cfg.db.GetPathComponents(bookIdStruct.BookId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error", err)
		return
	}

	oldPath, newPath, err := fileManagement.MoveFiles(downloadDir, cfg.downloadsPath, bookDir, cfg.libraryPath, authorDir, seriesDir)
	if err != nil {
		if os.IsExist(err) {
			respondWithError(w, http.StatusConflict, "The library already has files at the book's location", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Something went wrong moving files", err)
		return
	}

	book, err := cfg.db.AssociateBookAndDownload(bookIdStruct.BookId, downloadId, authorDir, seriesDir, bookDir)
	if err != nil {
		log.Println(err)
		err = fileManagement.MoveFilesWithPaths(newPath, oldPath)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to associate the book and files, and failed to move files back from the library to downloads", err)
			return
		}

		err := cfg.db.DeleteBookFilesFromDatabase(bookIdStruct.BookId)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to associate the book and files. Files have been returned to downloads. Failed to remove file paths from book", err)
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Failed to associate the book and files. Files have been returned to downloads", err)
		return
	}

	if !bookIdStruct.UseDownloadedCover {

		handleCoverReplacement := func() {
			fmt.Println()
			log.Println("Replacing the downloaded cover with the metadata cover")

			newCoverPath := path.Join(newPath, "cover.jpg")
			if _, err := os.Stat(newCoverPath); err == nil {
				err = os.Rename(newCoverPath, path.Join(newPath, "old-cover.jpg"))
				if err != nil {
					log.Println("Failed to rename existing cover")
					return
				}
			}

			metadataCoverPath := path.Join(cfg.metadataPath, bookIdStruct.BookId.String()+".jpg")

			if _, err := os.Stat(metadataCoverPath); err != nil {
				log.Println("Couldn't use the metadata cover because it doesn't exist")
				return
			}

			_, _, err = cfg.db.UpdateBookCover(bookIdStruct.BookId, "jpg")
			if err != nil {
				log.Println(err)
				return
			}

			err = fileManagement.MoveFilesWithPaths(metadataCoverPath, newCoverPath)
			if err != nil {
				log.Println("Failed to move metadata cover to the library")
			}

		}

		handleCoverReplacement()
		fmt.Println()
	}

	err = fileManagement.CreateMetadataFile(metadata.MetadataFileFromBook(book), path.Join(newPath, "metadata.json"))
	if err != nil {
		log.Println("failed to create metadata file:", err)
	}

	book.Files.Prepend(cfg.libraryName)

	respondWithJson(w, http.StatusOK, book)
}

func (cfg *apiConfig) handlerGetDownloadCover(id uuid.UUID, w http.ResponseWriter, r *http.Request) {
	download, err := cfg.db.GetDownload(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error", err)
		return
	}
	if download == nil {
		respondWithError(w, http.StatusBadRequest, "Download not found", err)
		return
	}

	coverPath := path.Join(cfg.downloadsPath, *download.Files.Root, *download.Files.Cover)
	log.Println("Serving download cover from", coverPath)

	http.ServeFile(w, r, coverPath)
}
