package fileManagement

import (
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func CopyFiles(oldPath, newPath string) error {
	return moveCopyFiles(oldPath, newPath, false)
}

func MoveFiles(oldPath, newPath string) error {
	return moveCopyFiles(oldPath, newPath, true)
}

func MoveOrCopyFiles(oldPath, newPath string, isMoving bool) error {
	return moveCopyFiles(oldPath, newPath, isMoving)
}

func moveCopyFiles(oldPath, newPath string, deleteOld bool) error {

	oldPath = filepath.Clean(oldPath)
	newPath = filepath.Clean(newPath)

	err := os.MkdirAll(filepath.Dir(newPath), 0755)
	if err != nil {
		return err
	}

	// If we aren't copying, we can try moving the file directly. If it fails then we're in an environment where streaming is required
	if deleteOld {
		err = os.Rename(oldPath, newPath)
		if err == nil {
			return nil
		}
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

				return streamFile(path, targetPath, deleteOld)
			})
			if err != nil {
				return err
			}

			if deleteOld {
				err = os.Remove(oldPath)
				if err != nil {
					log.Println("Failed to clean up old directory at \"", oldPath, "\"")
					return err
				}
			}

		} else {
			err = streamFile(oldPath, newPath, deleteOld)
		}
		if err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

func streamFile(oldPath, newPath string, removeOldFile bool) error {
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

	if removeOldFile {
		err = os.Remove(oldPath)
		if err != nil {
			log.Println("Failed to remove old file after copying from \"", oldPath, "\" to \"", newPath, "\"")
			return err
		}
	}

	return nil
}
