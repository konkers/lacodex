package ingest

import (
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

var intermediatePrefix = ""

func asRGBA(src image.Image) *image.RGBA {
	srcBounds := src.Bounds()
	destBounds := image.Rect(0, 0, srcBounds.Dx(), srcBounds.Dy())
	img := image.NewRGBA(destBounds)
	draw.Draw(img, destBounds, src, srcBounds.Min, draw.Src)
	return img
}

func writeIntermediateText(tag string, text string) {
	if intermediatePrefix == "" {
		return
	}

	os.MkdirAll("intermediates", 0755)
	fileName := filepath.Join("intermediates", fmt.Sprintf("%s-%s.txt", intermediatePrefix, tag))
	writer, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Failed to write intermediate file %s: %v\n", fileName, err)
		return
	}
	defer writer.Close()

	writer.WriteString(text)
}

func writeIntermediateJson(tag string, obj interface{}) {
	if intermediatePrefix == "" {
		return
	}

	os.MkdirAll("intermediates", 0755)
	fileName := filepath.Join("intermediates", fmt.Sprintf("%s-%s.json", intermediatePrefix, tag))
	writer, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Failed to write intermediate file %s: %v\n", fileName, err)
		return
	}
	defer writer.Close()

	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		fmt.Printf("Failed to encode intermediate %s: %v\n", fileName, err)
		return
	}
	writer.Write(b)
}

func writeIntermediateImg(tag string, img image.Image) {
	if intermediatePrefix == "" {
		return
	}

	os.MkdirAll("intermediates", 0755)

	imgName := filepath.Join("intermediates", fmt.Sprintf("%s-%s.png", intermediatePrefix, tag))
	writer, err := os.OpenFile(imgName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Failed to write intermediate image %s: %v\n", imgName, err)
		return
	}

	err = png.Encode(writer, img)
	if err != nil {
		fmt.Printf("Failed to endode intermediate image %s: %v\n", imgName, err)
		return
	}

}
