package fileManagement

import (
	"encoding/json"
	"os"

	"github.com/Ethanol2/book-organizer/internal/metadata"
)

func CreateMetadataFile(metadata metadata.MetadataFile, path string) error {

	jsonBytes, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	return os.WriteFile(path, jsonBytes, 0644)
}
