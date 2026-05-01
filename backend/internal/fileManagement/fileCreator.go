package fileManagement

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // Registers PNG decoder
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

func CreateMetadataFile(metadata MetadataFile, dirPath string) error {

	jsonBytes, err := json.MarshalIndent(metadata, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(path.Join(dirPath, "metadata.json"), jsonBytes, 0644)
}

func DownloadTempFile(url string) (*os.File, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		log.Println("Failed to decode")
		return nil, err
	}

	tmp, err := os.CreateTemp("", "bookOrg-*.jpg")
	if err != nil {
		return nil, err
	}

	err = jpeg.Encode(tmp, img, &jpeg.Options{Quality: 90})
	if err != nil {
		log.Println("Failed to encode")
		return nil, err
	}

	return tmp, nil
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

func HandleTestCover(testPath, metadataPath, imgSource string) error {

	// Check the cover exists in the test folder
	if _, err := os.Stat(testPath); err == nil {

		data, err := os.ReadFile(testPath)
		if err != nil {
			log.Println("Error trying to read the test cover in the test covers folder \"", testPath, "\"")
			return err
		}

		err = os.WriteFile(metadataPath, data, 0644)
		if err != nil {
			log.Println("Error trying to copy the test cover to the metadata folder =>", err)
			return err
		}

		log.Println("Using the saved test cover")
		fmt.Println()
		return nil
	}

	log.Println("Downloading cover from \"", imgSource, "\"")
	var coverFile *os.File
	coverFile, err := DownloadTempFile(imgSource)
	if err != nil {
		return err
	}
	defer coverFile.Close()

	log.Println("Moving file to \"", testPath, "\"")
	err = MoveFilesWithPaths(coverFile.Name(), testPath)
	if err != nil {
		return err
	}

	log.Println("Reading file to copy it to the metadata folder")
	coverFileContent, err := os.ReadFile(testPath)
	if err != nil {
		return err
	}

	log.Println("Writing copy to the metadata folder \"", metadataPath, "\"")
	err = os.WriteFile(metadataPath, coverFileContent, 0644)
	if err != nil {
		return err
	}

	return nil
}

func CreateTestDirectory(dirPath, pathPrefix string, metadata *MetadataFile, coverSourcePath string, fakeAudioCount, fakeTextCount int) (Files, error) {

	log.Println("Creating directory at \"", dirPath, "\"")
	err := CreateDirectory(path.Join(pathPrefix, dirPath))
	if err != nil {
		return Files{}, err
	}

	files := Files{
		Root: &dirPath,
	}

	if metadata != nil {
		log.Println("Writing metadata file")
		err = CreateMetadataFile(*metadata, path.Join(pathPrefix, dirPath))
		if err != nil {
			return Files{}, err
		}

		files.HasMetadata = true
	}

	if coverSourcePath != "" {
		log.Println("Reading cover from metadata \"", coverSourcePath, "\"")
		coverFileContent, err := os.ReadFile(coverSourcePath)
		if err != nil {
			return Files{}, err
		}

		coverPath := path.Join(dirPath, "cover.jpg")

		log.Println("Writing copy to the new folder")
		err = os.WriteFile(path.Join(pathPrefix, coverPath), coverFileContent, 0644)
		if err != nil {
			return Files{}, err
		}

		files.Cover = &coverPath
	}

	if fakeAudioCount > 0 {

		fileList := []string{}

		// Logic provided by ai
		createFakeM4b := func(filePath string, size int64) {
			f, err := os.Create(filePath)
			if err != nil {
				fmt.Println("Error creating file:", err)
				return
			}
			defer f.Close()

			// RIFF chunk
			f.Write([]byte("RIFF"))
			binary.Write(f, binary.LittleEndian, size)
			// WAVE chunk
			f.Write([]byte("WAVE"))
			binary.Write(f, binary.LittleEndian, size)
			// fmt chunk
			f.Write([]byte("fmt "))
			binary.Write(f, binary.LittleEndian, size)
			// data chunk
			f.Write([]byte("data"))
			binary.Write(f, binary.LittleEndian, size)
			// MOOV chunk
			f.Write([]byte("MOOV"))
			binary.Write(f, binary.LittleEndian, size)
			// mvhd chunk
			f.Write([]byte("mvhd"))
			binary.Write(f, binary.LittleEndian, size)
			// elst chunk
			f.Write([]byte("elst"))
			binary.Write(f, binary.LittleEndian, size)
			// meta chunk
			f.Write([]byte("meta"))
			binary.Write(f, binary.LittleEndian, size)

			// Fill the data chunk with random bytes
			randomBytes := make([]byte, size-28)
			_, err = f.Write(randomBytes)
			if err != nil {
				fmt.Println("Error writing random bytes:", err)
				return
			}
		}

		// Random char added at end so it can be chopped off
		fileName := "chapter-%d.m4b#"
		if fakeAudioCount == 1 {
			fileName = path.Base(dirPath) + ".m4b%d"
		}

		for i := range fakeAudioCount {
			aPath := path.Join(dirPath, fmt.Sprintf(fileName, i)[:len(fileName)-2])
			fileList = append(fileList, aPath)
			createFakeM4b(path.Join(pathPrefix, aPath), int64(1000000))
		}

		files.AudioFiles = &fileList
	}

	if fakeTextCount > 0 {

		fileList := []string{}

		// logic provided by ai
		createFakeEpub := func(numPages int) ([]byte, error) {
			// Create the fake content
			content := make([][]byte, numPages)
			for i := 0; i < numPages; i++ {
				pageContent := fmt.Sprintf("This is page %d of the fake EPUB.", i+1)
				content[i] = []byte(pageContent)
			}

			// Create the EPUB structure
			epubBytes := []byte(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="2.0">
  <metadata>
    <dc:title>My Fake EPUB</dc:title>
    <dc:creator>John Doe</dc:creator>
    <meta name="cover" content="cover"/>
    <meta name="dtb:uid" content="urn:uuid:12345678"/>
  </metadata>
  <manifest>
    <item id="ncx" href="toc.ncx" media-type="application/x-dtb+zip"/>
    <item id="cover" href="cover.xhtml" media-type="application/xhtml+xml"/>
`)
			for i := 0; i < numPages; i++ {
				epubBytes = append(epubBytes, []byte(fmt.Sprintf(`<item id="page%d" href="page%d.xhtml" media-type="application/xhtml+xml"/>\n`, i+1, i+1))...)
			}
			tmp := []byte(`</manifest>
  <spine toc="ncx">`)
			epubBytes = append(epubBytes, tmp...)
			for i := 0; i < numPages; i++ {
				epubBytes = append(epubBytes, []byte(fmt.Sprintf(`<itemref idref="page%d" linear="yes"/>\n`, i+1))...)
			}
			tmp = []byte(`</spine>
</package>`)
			epubBytes = append(epubBytes, tmp...)

			// Append the content to the EPUB structure
			for _, pageBytes := range content {
				epubBytes = append(epubBytes, pageBytes...)
			}

			return epubBytes, nil
		}

		log.Println("Creating a fake epub file")
		body, err := createFakeEpub(2)
		if err != nil {
			return Files{}, err
		}

		// Random char added at end so it can be chopped off
		textName := "chapter-%d.epub#"
		if fakeTextCount == 1 {
			textName = path.Base(dirPath) + ".epub%d"
		}

		log.Println("Writing fake epub to disk", fakeTextCount, "times")
		for i := range fakeTextCount {

			fPath := path.Join(dirPath, fmt.Sprintf(textName, i)[:len(textName)-2])
			fileList = append(fileList, fPath)

			err = os.WriteFile(path.Join(pathPrefix, fPath), body, 0644)
			if err != nil {
				log.Println(err)
				continue
			}
		}

		files.TextFiles = &fileList
	}

	return files, nil
}
