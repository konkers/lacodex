package ingest

import (
	"fmt"
	"image"
	_ "image/png" // Pull in png decoder.
	"os"
)

var referenceImageCache = map[string]image.Image{}

func getReferenceImage(name string) (image.Image, error) {
	img, ok := referenceImageCache[name]
	if ok {
		return img, nil
	}

	imgFile := fmt.Sprintf("reference/%s.png", name)
	reader, err := os.Open(imgFile)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	img, _, err = image.Decode(reader)
	if err != nil {
		return nil, err
	}

	referenceImageCache[name] = img
	return img, nil
}
