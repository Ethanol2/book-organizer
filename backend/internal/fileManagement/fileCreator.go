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

	tmp, err := os.CreateTemp("", "bookOrg-*"+ext)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(tmp, resp.Body)
	if err != nil {
		return nil, err
	}

	return tmp, err
}

func CreateTempFileFromRequest(r *http.Request) (*os.File, error) {
	contentType := r.Header.Get("Content-Type")
	ext := ""
	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	case "image/gif":
		ext = ".gif"
	default:
		ext = ".jpg" // fallback
	}

	tmp, err := os.CreateTemp("", "bookOrg-*"+ext)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(tmp, r.Body)
	if err != nil {
		return nil, err
	}

	return tmp, nil
}
