package database

import "encoding/json"

type BookFiles struct {
	AudioFiles *FileList `json:"audio_files"`
	TextFiles  *FileList `json:"text_files"`
	Cover      *string   `json:"cover"`
}
type FileList struct {
	Files []string `json:"files"`
}

func (files BookFiles) ToJson() (string, string, string, error) {

	audioBytes, err := json.Marshal(files.AudioFiles)
	if err != nil {
		return "", "", "", err
	}
	textBytes, err := json.Marshal(files.TextFiles)
	if err != nil {
		return "", "", "", err
	}

	return string(audioBytes), string(textBytes), *files.Cover, nil
}

func (files BookFiles) ParseAudioJson(audioJson string) error {
	err := json.Unmarshal([]byte(audioJson), &files.AudioFiles)
	if err != nil {
		return err
	}
	return nil
}

func (files BookFiles) ParseTextJson(textJson string) error {
	err := json.Unmarshal([]byte(textJson), &files.TextFiles)
	if err != nil {
		return err
	}
	return nil
}
