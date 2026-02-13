package fileManagement

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Ethanol2/book-organizer/internal/metadata"
)

func CreateMetadataFile(metadata metadata.MetadataFile, path string) error {

	jsonBytes, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	return os.WriteFile(path, jsonBytes, 0644)
}

func DownloadTempFile(url string) (*os.File, error) {

	ext := filepath.Ext(url)
	if ext == "" {
		ext = ".jpg"
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp("", "cover-*"+ext)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(tmp, resp.Body)
	if err != nil {
		return nil, err
	}

	return tmp, err
}
