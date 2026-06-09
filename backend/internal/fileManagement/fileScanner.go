package fileManagement

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Scanner struct {
	Frequency    time.Duration
	Directory    string
	DepthLimit   int
	IncludeEmpty bool
	running      bool

	AddHandler    func([]Files) error
	DeleteHandler func(uuid.UUID) error
	UpdateHandler func(uuid.UUID, Files) error
	GetExisting   func() ([]uuid.UUID, []string, error)
}

type FileType int

const (
	Audio FileType = iota
	Text
	Image
	Metadata
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
					log.Println("Scan error =>", err)
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

	if stat, err := os.Stat(scan.Directory); os.IsNotExist(err) {
		return fmt.Errorf("scan directory does not exist: %s", scan.Directory)
	} else if !stat.IsDir() {
		return fmt.Errorf("scan path points to a file: %s", scan.Directory)
	}

	filesList := []Files{}

	var scanFunc func(string, string, int) error
	scanFunc = func(path, relativePath string, depth int) error {

		if depth >= scan.DepthLimit || slices.Contains(toIgnore, relativePath) {
			return nil
		}

		dirContents, err := getFolderContents(path, relativePath)
		if err != nil {
			return err
		}

		if len(*dirContents.Directories) > 0 {
			for _, subDir := range *dirContents.Directories {

				err = scanFunc(
					filepath.Join(path, subDir),
					filepath.Join(relativePath, subDir),
					depth+1,
				)
				if err != nil {
					log.Println(err)
					continue
				}
			}
		} else {
			isEmpty := dirContents.HasNoFiles()

			if (isEmpty && scan.IncludeEmpty) || !isEmpty {
				filesList = append(filesList, dirContents)
			}
		}

		return nil
	}

	err := scanFunc(scan.Directory, "", 0)
	if err != nil {
		return err
	}

	err = scan.AddHandler(filesList)
	if err != nil {
		return err
	}

	return nil
}

func (scan *Scanner) ScanExisting(ids []uuid.UUID, dirs []string) error {

	for i, dir := range dirs {

		path := filepath.Join(scan.Directory, dir)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			err = scan.DeleteHandler(ids[i])
			if err != nil {
				log.Println(err)
			}
			log.Println("\"", dir, "\" was not found")
			continue
		}

		newFiles, err := getFolderContents(filepath.Join(scan.Directory, dir), dir)
		if err != nil {
			log.Println(err)
			continue
		}

		if newFiles.HasNoFiles() && !scan.IncludeEmpty {
			err = scan.DeleteHandler(ids[i])
			if err != nil {
				log.Println(err)
			}
			log.Println("\"", dir, "\" was found but is empty")
		} else {
			err = scan.UpdateHandler(ids[i], newFiles)
			if err != nil {
				return err
			}
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

	case ".json":
		if filename == "metadata.json" {
			return Metadata
		}
	}

	return Other
}

func getFolderContents(path, relativePath string) (Files, error) {

	hasMD := false
	var audio []string
	var text []string
	var images []string
	var others []string
	var dirs []string

	bookItems, err := os.ReadDir(path)
	if err != nil {
		log.Println(err)
		return Files{}, nil
	}

	for _, item := range bookItems {

		if item.Type().IsDir() {
			dirs = append(dirs, item.Name())
			continue
		}

		fileType := getFileType(item.Name())
		//log.Println(item.Name(), "->", fileType)

		switch fileType {

		case Image:
			images = append(images, filepath.Join(relativePath, item.Name()))

		case Text:
			text = append(text, filepath.Join(relativePath, item.Name()))

		case Audio:
			audio = append(audio, filepath.Join(relativePath, item.Name()))

		case Metadata:
			hasMD = true

		case Other:
			others = append(others, filepath.Join(relativePath, item.Name()))
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
		Root:        &relativePath,
		AudioFiles:  &audio,
		TextFiles:   &text,
		OtherFiles:  &others,
		Cover:       cover,
		HasMetadata: hasMD,

		Directories: &dirs,
	}, nil
}
