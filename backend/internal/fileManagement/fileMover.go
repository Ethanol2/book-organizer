package fileManagement

import (
	"os"
	"path"
)

func MoveFiles(targetDir, fromDir, toDir, author, series string) (string, string, error) {

	authorDir := path.Join(toDir, author)
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

	fromPath := path.Join(fromDir, targetDir)
	toPath := path.Join(seriesDir, targetDir)

	err := os.Rename(fromPath, toPath)
	if err != nil {
		return "", "", err
	}

	return fromPath, toPath, nil
}

func MoveFilesWithPaths(fromPath, toPath string) error {
	err := os.Rename(fromPath, toPath)
	if err != nil {
		return err
	}
	return nil
}
