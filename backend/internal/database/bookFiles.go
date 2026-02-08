package database

import (
	"encoding/json"
	"path"
)

type BookFiles struct {
	Root       string    `json:"root"`
	AudioFiles *[]string `json:"audio_files"`
	TextFiles  *[]string `json:"text_files"`
	Cover      *string   `json:"cover"`
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

func (files *BookFiles) ParseAudioJson(audioJson string) error {
	err := json.Unmarshal([]byte(audioJson), &files.AudioFiles)
	if err != nil {
		return err
	}
	return nil
}

func (files *BookFiles) ParseTextJson(textJson string) error {
	err := json.Unmarshal([]byte(textJson), &files.TextFiles)
	if err != nil {
		return err
	}
	return nil
}

func (files *BookFiles) Prepend(p string) {

	prepend := func(items []string) *[]string {
		for i := range items {
			items[i] = path.Join(p, items[i])
		}
		return &items
	}

	if files.AudioFiles != nil {
		files.AudioFiles = prepend(*files.AudioFiles)
	}
	if files.TextFiles != nil {
		files.TextFiles = prepend(*files.TextFiles)
	}
	if files.Cover != nil {
		cover := path.Join(p, *files.Cover)
		files.Cover = &cover
	}

	files.Root = path.Join(p, files.Root)
}
