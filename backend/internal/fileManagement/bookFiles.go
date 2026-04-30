package fileManagement

import (
	"encoding/json"
	"path"
)

type Files struct {
	Root       *string       `json:"root"`
	AudioFiles *[]string     `json:"audio_files"`
	TextFiles  *[]string     `json:"text_files"`
	Cover      *string       `json:"cover"`
	Metadata   *MetadataFile `json:"metadata,omitempty"`
}

// Matches AudioBookshelf's metadata format
type MetadataFile struct {
	Tags     []string `json:"tags"`
	Chapters []struct {
		ID    int     `json:"id"`
		Start int     `json:"start"`
		End   float64 `json:"end"`
		Title string  `json:"title"`
	} `json:"chapters,omitempty"`
	Title         string   `json:"title"`
	Subtitle      *string  `json:"subtitle,omitempty"`
	Authors       []string `json:"authors"`
	Narrators     []string `json:"narrators"`
	Series        []string `json:"series"`
	Genres        []string `json:"genres"`
	PublishedYear string   `json:"publishedYear"`
	PublishedDate *string  `json:"publishedDate"`
	Publisher     string   `json:"publisher"`
	Description   string   `json:"description"`
	Isbn          string   `json:"isbn"`
	Asin          string   `json:"asin"`
	Language      string   `json:"language"`
	Explicit      bool     `json:"explicit,omitempty"`
	Abridged      bool     `json:"abridged,omitempty"`
}

func (files Files) FileListsToJson() (string, string, error) {

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

func (files *Files) ParseAudioJson(audioJson string) error {
	err := json.Unmarshal([]byte(audioJson), &files.AudioFiles)
	if err != nil {
		return err
	}
	return nil
}

func (files *Files) ParseTextJson(textJson string) error {
	err := json.Unmarshal([]byte(textJson), &files.TextFiles)
	if err != nil {
		return err
	}
	return nil
}

func (files *Files) Prepend(p string) {

	prepend := func(items []string) *[]string {
		for i := range items {
			items[i] = path.Join(p, items[i])
		}
		return &items
	}

	files.applyModifier(prepend)
}

func (files *Files) UpdateDirectory(dir string) {

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

func (files *Files) applyModifier(mod func([]string) *[]string) {
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
