package fileManagement

import (
	"os"
	"path"
)

func MoveFiles(targetDir, oldDir, newDir, author, series string) (string, string, error) {

	authorDir := path.Join(newDir, author)
	if _, err := os.Stat(authorDir); os.IsNotExist(err) {
		err = os.Mkdir(authorDir, os.ModePerm)
		if err != nil {
			return "", "", err
		}
	}

	seriesDir := path.Join(authorDir, series)
	if series != "" {
		if _, err := os.Stat(seriesDir); os.IsNotExist(err) {
			err = os.Mkdir(seriesDir, os.ModePerm)
			if err != nil {
				return "", "", err
			}
		}
	}

	oldPath := path.Join(oldDir, targetDir)
	newPath := path.Join(seriesDir, targetDir)

	err := os.Rename(oldPath, newPath)
	if err != nil {
		return "", "", err
	}

	return oldPath, newPath, nil
}

func MoveFilesWithPaths(oldPath, newPath string) error {
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return err
	}
	return nil
}
