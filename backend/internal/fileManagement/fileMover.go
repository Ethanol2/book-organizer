package fileManagement

import (
	"os"
	"path/filepath"
)

func MoveFiles(oldPath, newPath string) error {

	err := os.MkdirAll(filepath.Dir(newPath), 0755)
	if err != nil {
		return err
	}

	// If this succeeds then we don't have to stream the file
	err = os.Rename(oldPath, newPath)
	if err == nil {
		return nil
	}

	return nil
}
