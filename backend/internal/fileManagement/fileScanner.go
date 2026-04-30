package fileManagement

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

	"github.com/google/uuid"
)

type Scanner struct {
	Frequency time.Duration
	Directory string
	running   bool

	AddHandler    func([]Files) error
	DeleteHandler func(uuid.UUID) error
	UpdateHandler func(map[uuid.UUID]Files) error
	GetExisting   func() ([]uuid.UUID, []string, error)
}

type FileType int

const (
	Audio FileType = iota
	Text
	Image
	Other
)

func (scan *Scanner) Start(ctx context.Context) error {

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

				err := scan.Scan()
				if err != nil {
					log.Println(err)
				}

				time.Sleep(scan.Frequency)
			}
		}

	}(ctx)

	return nil
}

func (scan *Scanner) Scan() error {

	//log.Println("Scanning...")

	ids, dirs, err := scan.GetExisting()
	if err != nil {
		return err
	}

	err = scan.ScanExisting(ids, dirs)
	if err != nil {
		return err
	}

	err = scan.ScanNew(dirs)
	if err != nil {
		return err
	}

	return nil
}

func (scan *Scanner) ScanNew(toIgnore []string) error {

	if _, err := os.Stat(scan.Directory); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist -> %s", scan.Directory)
	}

	dirItems, err := os.ReadDir(scan.Directory)
	if err != nil {
		return err
	}

	downloads := []Files{}

	for _, item := range dirItems {

		if !item.Type().IsDir() || slices.Contains(toIgnore, item.Name()) {
			continue
		}

		files, err := getFiles(scan.Directory, item.Name())
		if err != nil {
			log.Println(err)
			continue
		}

		downloads = append(downloads, files)
	}

	err = scan.AddHandler(downloads)
	if err != nil {
		return err
	}

	return nil
}

func (scan *Scanner) ScanExisting(ids []uuid.UUID, dirs []string) error {

	files := map[uuid.UUID]Files{}

	for i, dir := range dirs {

		path := path.Join(scan.Directory, dir)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			err = scan.DeleteHandler(ids[i])
			if err != nil {
				log.Println(err)
			}
			log.Println("Download \"", dir, "\" was not found and removed from the database")
			continue
		}

		newFiles, err := getFiles(scan.Directory, dir)
		if err != nil {
			log.Println(err)
			continue
		}

		files[ids[i]] = newFiles
	}

	err := scan.UpdateHandler(files)
	if err != nil {
		return err
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

func getFiles(root, folder string) (Files, error) {

	p := path.Join(root, folder)

	var audio []string
	var text []string
	var images []string

	bookItems, err := os.ReadDir(p)
	if err != nil {
		log.Println(err)
		return Files{}, nil
	}

	for _, item := range bookItems {

		fileType := getFileType(item.Name())
		//log.Println(item.Name(), "->", fileType)

		switch fileType {

		case Image:
			images = append(images, path.Join(folder, item.Name()))

		case Text:
			text = append(text, path.Join(folder, item.Name()))

		case Audio:
			audio = append(audio, path.Join(folder, item.Name()))
		}
	}

	var cover *string

	if len(images) > 1 {
		coverStr := ""
		for _, img := range images {
			if strings.Contains(strings.ToLower(img), "cover") {
				coverStr = img
				break
			}
		}

		if coverStr == "" {
			coverStr = images[0]
		}
		cover = &coverStr

	} else if len(images) == 1 {
		cover = &images[0]
	} else {
		cover = nil
	}

	return Files{
		Root:       &folder,
		AudioFiles: &audio,
		TextFiles:  &text,
		Cover:      cover,
	}, nil
}
