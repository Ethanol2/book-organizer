package fileManagement

import (
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func MoveFiles(oldPath, newPath string) error {

	oldPath = filepath.Clean(oldPath)
	newPath = filepath.Clean(newPath)

	err := os.MkdirAll(filepath.Dir(newPath), 0755)
	if err != nil {
		return err
	}

	// If this succeeds then we don't have to stream the file
	err = os.Rename(oldPath, newPath)
	if err == nil {
		return nil
	}

	if stat, err := os.Stat(oldPath); err == nil {
		if stat.IsDir() {

			err = filepath.Walk(oldPath, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}

				relPath, err := filepath.Rel(oldPath, path)
				if err != nil {
					return err
				}

				targetPath := filepath.Join(newPath, relPath)

				if info.IsDir() {
					return os.MkdirAll(targetPath, info.Mode())
				}

				return moveFile(path, targetPath)
			})
			if err != nil {
				return err
			}

			err = os.Remove(oldPath)
			if err != nil {
				log.Println("Failed to clean up old directory at \"", oldPath, "\"")
				return err
			}

		} else {
			err = moveFile(oldPath, newPath)
		}
		if err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

func moveFile(oldPath, newPath string) error {
	srcFile, err := os.Open(oldPath)
	if err != nil {
		log.Println("Failed to open file at \"", oldPath, "\"")
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(newPath)
	if err != nil {
		log.Println("Failed to create destination file at \"", newPath, "\"")
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		log.Println("Failed to stream data from \"", oldPath, "\" to \"", newPath, "\"")
		return err
	}

	if err = os.Remove(oldPath); err != nil {
		log.Println("Failed to remove old file after copying from \"", oldPath, "\" to \"", newPath, "\"")
		return err
	}

	return nil
}
