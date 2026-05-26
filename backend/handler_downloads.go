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

	"github.com/Ethanol2/book-organizer/internal/env"
	"github.com/Ethanol2/book-organizer/internal/fileManagement"
	"github.com/Ethanol2/book-organizer/internal/metadata"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerGetDownloads(w http.ResponseWriter, r *http.Request) {

	log.Println("Fetching downloads")
	downloads, err := cfg.db.GetDownloads()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	for i := range downloads {
		downloads[i].Files.Prepend(cfg.downloadsName)
	}

	respondWithJson(w, http.StatusOK, downloads)
}

func (cfg *apiConfig) handlerGetDownload(id uuid.UUID, w http.ResponseWriter, r *http.Request) {

	download, err := cfg.db.GetDownload(id)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, NotFoundError, err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	download.Files.Prepend(cfg.downloadsName)

	respondWithJson(w, http.StatusOK, download)
}

func (cfg *apiConfig) handlerAssociateDownloadToBook(downloadId uuid.UUID, w http.ResponseWriter, r *http.Request) {

	downloadDir, err := cfg.db.GetDownloadDir(downloadId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusBadRequest, "Download "+NotFoundError, err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	var bookIdStruct struct {
		BookId             uuid.UUID `json:"book_id"`
		UseDownloadedCover bool      `json:"use_downloaded_cover"`
	}
	err = json.NewDecoder(r.Body).Decode(&bookIdStruct)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, BodyDecodeError, err)
		return
	}

	bookExists, err := cfg.db.CheckBookExistsID(bookIdStruct.BookId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}
	if !bookExists {
		respondWithError(w, http.StatusNotFound, "Book "+NotFoundError, err)
		return
	}

	bookPath, err := cfg.db.GetBookPath(bookIdStruct.BookId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	oldPath := path.Join(cfg.Env.Get(env.DownloadsPath), downloadDir)
	newPath := path.Join(cfg.Env.Get(env.LibraryPath), bookPath)

	err = fileManagement.MoveFiles(oldPath, newPath)
	if err != nil {
		if os.IsExist(err) {
			respondWithError(w, http.StatusConflict, "The library already has files at the book's location", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, FileMoveError, err)
		return
	}

	book, err := cfg.db.AssociateBookAndDownload(bookIdStruct.BookId, downloadId, bookPath)
	if err != nil {
		log.Println(err)
		err = fileManagement.MoveFiles(newPath, oldPath)
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

			err = fileManagement.MoveFiles(metadataCoverPath, newCoverPath)
			if err != nil {
				log.Println("Failed to move metadata cover to the library")
			}

		}

		handleCoverReplacement()
		fmt.Println()
	}

	err = fileManagement.CreateMetadataFile(*metadata.BookToMetadata(book), newPath)
	if err != nil {
		log.Println("failed to create metadata file:", err)
	}

	book.Files.Prepend(cfg.libraryName)
	respondWithJson(w, http.StatusOK, book)
}

func (cfg *apiConfig) handlerGetDownloadCover(id uuid.UUID, w http.ResponseWriter, r *http.Request) {
	download, err := cfg.db.GetDownload(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusBadRequest, "Download "+NotFoundError, err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}
	if download == nil {
		respondWithError(w, http.StatusBadRequest, NotFoundError, err)
		return
	}

	coverPath := path.Join(cfg.Env.Get(env.DownloadsPath), *download.Files.Root, *download.Files.Cover)
	log.Println("Serving download cover from", coverPath)

	http.ServeFile(w, r, coverPath)
}
func (cfg *apiConfig) handlerGetDownloadMetadata(id uuid.UUID, w http.ResponseWriter, r *http.Request) {

	download, err := cfg.db.GetDownload(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusBadRequest, "Download "+NotFoundError, err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, DatabaseError, err)
		return
	}

	if !download.Files.HasMetadata {
		respondWithError(w, http.StatusNoContent, NotFoundError, err)
		return
	}

	metadataFile, err := os.Open(path.Join(cfg.Env.Get(env.DownloadsPath), *download.Files.Root, "metadata.json"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, FileOpenError, err)
		return
	}
	defer metadataFile.Close()

	var metadataContent fileManagement.MetadataFile
	err = json.NewDecoder(metadataFile).Decode(&metadataContent)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, FileOpenError, err)
		return
	}

	metadataData := metadata.MetadataToBookParams(metadataContent)
	coverPath := path.Join(cfg.downloadsName, *download.Files.Cover)
	metadataData.Cover = &coverPath

	respondWithJson(w, http.StatusOK, metadataData)
}
