package database

import (
	"encoding/json"
	"path"
)

type BookFiles struct {
	Root       *string   `json:"root"`
	AudioFiles *[]string `json:"audio_files"`
	TextFiles  *[]string `json:"text_files"`
	Cover      *string   `json:"cover"`
}

func (files BookFiles) FileListsToJson() (string, string, error) {

	audioBytes, err := json.Marshal(files.AudioFiles)
	if err != nil {
		return "", "", err
	}
	textBytes, err := json.Marshal(files.TextFiles)
	if err != nil {
		return "", "", err
	}

	return string(audioBytes), string(textBytes), nil
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

	files.applyModifier(prepend)
}

func (files *BookFiles) ReplaceDirectory(dir string) {

	replace := func(items []string) *[]string {
		for i := range items {
			fileName := path.Base(items[i])
			items[i] = path.Join(dir, fileName)
		}
		return &items
	}

	files.applyModifier(replace)
	files.Root = &dir
}

func (files *BookFiles) applyModifier(mod func([]string) *[]string) {
	if files.AudioFiles != nil {
		files.AudioFiles = mod(*files.AudioFiles)
	}
	if files.TextFiles != nil {
		files.TextFiles = mod(*files.TextFiles)
	}
	if files.Cover != nil {
		arr := []string{*files.Cover}
		cover := (*mod(arr))[0]
		files.Cover = &cover
	}
	if files.Root != nil {
		arr := []string{*files.Root}
		root := (*mod(arr))[0]
		files.Root = &root
	}
}
