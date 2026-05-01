package fileManagement

import (
	"log"
	"os"
	"path"
)

func RemoveDirectoryContents(dirPath string) error {

	log.Println("Gettings files from \"", dirPath, "\"")
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	log.Println("Deleting files")
	for _, file := range files {
		err = os.RemoveAll(path.Join(dirPath, file.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}
