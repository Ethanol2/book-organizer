package fileScanner

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/Ethanol2/book-organizer/internal/database"
)

type Scanner struct {
	Frequency time.Duration
	Directory string
	running   bool
}

type FileType int

const (
	Audio FileType = iota
	Text
	Image
	Other
)

func CreateNew(scanFrequency time.Duration, dir string) Scanner {
	return Scanner{
		Frequency: scanFrequency,
		Directory: dir,
	}
}

func (scan *Scanner) Start(ctx context.Context, db *database.Client) error {

	if scan.running {
		return fmt.Errorf("scanner already running")
	}

	go func(ctx context.Context) {

		scan.running = true
		for {
			select {
			case <-ctx.Done():
				scan.running = false
				return

			default:

				err := scan.Scan(db)
				if err != nil {
					log.Println(err)
				}

				time.Sleep(scan.Frequency)
			}
		}

	}(ctx)

	return nil
}

func (scan *Scanner) Scan(db *database.Client) error {

	log.Println("Scanning...")

	err := scan.ScanExisting(db)
	if err != nil {
		return err
	}

	err = scan.ScanNew(db)
	if err != nil {
		return err
	}

	return nil
}

func (scan *Scanner) ScanNew(db *database.Client) error {

	if _, err := os.Stat(scan.Directory); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist -> %s", scan.Directory)
	}

	_, knownDirs, err := db.GetAllDownloadsIdsAndDirs()
	if err != nil {
		return err
	}

	dirItems, err := os.ReadDir(scan.Directory)
	if err != nil {
		return err
	}

	for _, book := range dirItems {

		if !book.Type().IsDir() || slices.Contains(knownDirs, book.Name()) {
			continue
		}

		files, err := getFiles(path.Join(scan.Directory, book.Name()))
		if err != nil {
			log.Println(err)
			continue
		}

		_, err = db.CreateDownload(book.Name(), book.Name(), files)
	}

	return nil
}

func (scan *Scanner) ScanExisting(db *database.Client) error {

	ids, dirs, err := db.GetAllDownloadsIdsAndDirs()
	if err != nil {
		return err
	}

	for i, dir := range dirs {

		path := path.Join(scan.Directory, dir)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			err = db.DeleteDownload(ids[i])
			if err != nil {
				log.Println(err)
			}
			log.Println("Book \"", dir, "\" was not found and removed from the database")
			continue
		}

		newFiles, err := getFiles(path)
		if err != nil {
			log.Println(err)
			continue
		}

		err = db.UpdateDownloadFiles(ids[i], newFiles)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

func getFileType(filename string) FileType {

	ext := strings.ToLower(filepath.Ext(filename))
	//log.Println(ext)

	switch ext {

	case ".m4b", ".aax", ".mp3", ".aa", ".wma", ".flac", ".wav", ".daisy":
		return Audio

	case ".png", ".jpg", ".jpeg":
		return Image

	case ".epub", ".pdf", ".azw3", ".kfx", ".azw", ".mobi", ".iba", ".lrf", ".lrx", ".fb2", ".djvu", ".lit", ".prc", ".pdb", ".cbz", ".cbr", ".txt", ".rtf", ".html", ".docx":
		return Text
	}

	return Other
}

func getFiles(dir string) (database.BookFiles, error) {

	var audio []string
	var text []string
	var images []string

	bookItems, err := os.ReadDir(dir)
	if err != nil {
		log.Println(err)
		return database.BookFiles{}, nil
	}

	for _, item := range bookItems {

		fileType := getFileType(item.Name())
		//log.Println(item.Name(), "->", fileType)

		switch fileType {

		case Image:
			images = append(images, item.Name())

		case Text:
			text = append(text, item.Name())

		case Audio:
			audio = append(audio, item.Name())
		}
	}

	cover := ""

	if len(images) > 1 {

		for _, img := range images {
			if strings.Contains("cover", strings.ToLower(img)) {
				cover = img
				break
			}
		}

		if cover == "" {
			cover = images[0]
		}

	} else if len(images) == 1 {
		cover = images[0]
	}

	return database.BookFiles{
		AudioFiles: database.FileList{Files: audio},
		TextFiles:  database.FileList{Files: text},
		Cover:      cover,
	}, nil

}
