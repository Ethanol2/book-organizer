package fileManagement

import (
	"os"
	"path"
)

func CreateDirectory(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func MoveFiles(oldDirName, oldDir, newDirName, newDir, author, series string) (string, string, error) {

	authorDir := path.Join(newDir, author)
	err := CreateDirectory(authorDir)
	if err != nil {
		return "", "", err
	}

	seriesDir := path.Join(authorDir, series)
	err = CreateDirectory(seriesDir)
	if err != nil {
		return "", "", err
	}

	oldPath := path.Join(oldDir, oldDirName)
	newPath := path.Join(seriesDir, newDirName)

	err = os.Rename(oldPath, newPath)
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

func DeleteFiles(path string) error {
	err := os.RemoveAll(path)
	return err
}
