package fileScanner

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
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

func (scanner *Scanner) Start(ctx context.Context, db *database.Client) error {

	if scanner.running {
		return fmt.Errorf("scanner already running")
	}

	go func(ctx context.Context) {

		scanner.running = true
		for {
			select {
			case <-ctx.Done():
				scanner.running = false
				return

			default:

				err := scanner.Scan(db)
				if err != nil {
					log.Println(err)
				}

				time.Sleep(scanner.Frequency)
			}
		}

	}(ctx)

	return nil
}

func (scanner *Scanner) Scan(db *database.Client) error {

	if _, err := os.Stat(scanner.Directory); os.IsNotExist(err) {
		return err // fmt.Errorf("directory does not exist -> %s", scanner.Directory)
	}

	dirItems, err := os.ReadDir(scanner.Directory)
	if err != nil {
		return err
	}

	for _, book := range dirItems {

		if !book.Type().IsDir() {
			continue
		}

		var audio []string
		var text []string
		var images []string

		bookItems, err := os.ReadDir(path.Join(scanner.Directory, book.Name()))
		if err != nil {
			log.Println(err)
			continue
		}

		for _, item := range bookItems {

			fileType := getFileType(item.Name())
			log.Println(item.Name(), "->", fileType)

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

		_, err = db.CreateDownload(book.Name(), book.Name(), cover, audio, text)
	}

	return nil
}

func getFileType(filename string) FileType {

	ext := strings.ToLower(filepath.Ext(filename))
	log.Println(ext)

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
