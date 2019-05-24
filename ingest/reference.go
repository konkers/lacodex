package ingest

import (
	"fmt"
	"image"
	_ "image/png" // Pull in png decoder.
	"os"
	"path/filepath"
)

var referenceImageCache = map[string]image.Image{}

func clearReferenceImageCache() {
	referenceImageCache = map[string]image.Image{}
}

func getReferenceImage(name string) (image.Image, error) {
	img, ok := referenceImageCache[name]
	if ok {
		return img, nil
	}

	execPath, err := os.Executable()
	if ok {
		return nil, err
	}
	execDir, _ := filepath.Split(execPath)

	paths := []string{
		"reference",
		filepath.Join("ingest", "reference"),
		filepath.Join(execDir, "..", "src", "github.com", "konkers",
			"lacodex", "ingest", "reference"),
	}

	imgName := fmt.Sprintf("%s.png", name)
	var reader *os.File
	for _, dir := range paths {
		imgFile := filepath.Join(dir, imgName)
		reader, err = os.Open(imgFile)
		if err == nil {
			break
		}
	}

	if reader == nil {
		return nil, fmt.Errorf("Can't find reference image for %s.  Tried: %#v",
			name, paths)
	}

	defer reader.Close()
	img, _, err = image.Decode(reader)
	if err != nil {
		return nil, err
	}

	referenceImageCache[name] = img
	return img, nil
}
