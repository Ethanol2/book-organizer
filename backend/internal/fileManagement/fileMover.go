package fileManagement

import (
	"fmt"
	"os"
)

func CreateDirectory(path string) error {
	if info, err := os.Stat(path); err != nil {
		err = os.Mkdir(path, os.ModePerm)
		if err != nil {
			return err
		}
	} else if !info.IsDir() {
		return fmt.Errorf("couldn't create folder at \"%s\" because the path already points to a file", path)
	}
	return nil
}

func MoveFiles(oldPath, newPath string) error {

	err := os.MkdirAll(newPath, 0755)
	if err != nil {
		return err
	}

	err = os.Rename(oldPath, newPath)
	if err != nil {
		return err
	}
	return nil
}

func DeleteFiles(path string) error {
	err := os.RemoveAll(path)
	return err
}
