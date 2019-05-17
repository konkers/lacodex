package ingest

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/anthonynsimon/bild/effect"
	"github.com/anthonynsimon/bild/transform"
	"github.com/otiai10/gosseract"
)

var intermediatePrefix = ""

const nativeWidth = 640
const nativeHeight = 480

const msxContentWidth = 604
const msxContentHeight = 412

func writeIntermediateText(tag string, text string) {
	if intermediatePrefix == "" {
		return
	}

	os.MkdirAll("intermediates", 0755)
	fileName := filepath.Join("intermediates", fmt.Sprintf("%s-%s.txt", intermediatePrefix, tag))
	writer, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Failed to write intermediate ie %s: %v ", fileName, err)
		return
	}
	defer writer.Close()

	writer.WriteString(text)
}

func writeIntermediateImg(tag string, img image.Image) {
	if intermediatePrefix == "" {
		return
	}

	os.MkdirAll("intermediates", 0755)

	imgName := filepath.Join("intermediates", fmt.Sprintf("%s-%s.png", intermediatePrefix, tag))
	writer, err := os.OpenFile(imgName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Failed to write intermediate image %s: %v ", imgName, err)
		return
	}

	err = png.Encode(writer, img)
	if err != nil {
		fmt.Printf("Failed to endode intermediate image %s: %v ", imgName, err)
		return
	}

}

func middleCrop(img image.Image, width int, height int) image.Image {
	bounds := img.Bounds()
	insetX := bounds.Min.X + (bounds.Dx()-width)/2
	insetY := bounds.Min.Y + (bounds.Dy()-height)/2
	cropRect := image.Rect(insetX, insetY, width+insetX, height+insetY)
	return transform.Crop(img, cropRect)
}

func cropGameImage(img image.Image) image.Image {
	// First calculate the scale and crop.
	bounds := img.Bounds()
	scale := bounds.Dy() / nativeHeight

	resizedImg := transform.Resize(img, bounds.Dx()/scale, bounds.Dy()/scale, transform.NearestNeighbor)
	writeIntermediateImg("resized", resizedImg)

	croppedGameImg := middleCrop(resizedImg, nativeWidth, nativeHeight)
	writeIntermediateImg("cropped-game", croppedGameImg)

	return croppedGameImg
}

// Takes a cropped game image.
func msxContent(img image.Image) image.Image {
	croppedContentImg := middleCrop(img, msxContentWidth, msxContentHeight)
	writeIntermediateImg("cropped-content", croppedContentImg)

	return croppedContentImg
}

// Takes a cropped msx image
func ocrPrep(img image.Image) image.Image {
	invertedImg := effect.Invert(img)
	writeIntermediateImg("ocr-inverted", invertedImg)

	greyImg := effect.Grayscale(invertedImg)
	writeIntermediateImg("ocr-greyscale", greyImg)

	return greyImg
}

func ocr(img image.Image) (string, error) {

	var b bytes.Buffer
	err := png.Encode(&b, img)
	if err != nil {
		return "", fmt.Errorf("Failed to endode image to png buffer: %v ", err)
	}

	client := gosseract.NewClient()
	defer client.Close()
	client.SetImageFromBytes(b.Bytes())
	text, err := client.Text()
	if err != nil {
		return "", err
	}
	writeIntermediateText("ocr", text)

	return text, nil
}
