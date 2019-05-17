package ingest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthonynsimon/bild/effect"
	"github.com/anthonynsimon/bild/transform"
	"github.com/konkers/lacodex/model"
	"github.com/otiai10/gosseract"
)

var intermediatePrefix = ""

const nativeWidth = 640
const nativeHeight = 480

const msxContentWidth = 604
const msxContentHeight = 412

var normalColor = color.RGBA{230, 232, 236, 255}
var blueColor = color.RGBA{100, 182, 227, 255}
var greenColor = color.RGBA{96, 229, 147, 255}

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
	writeIntermediateImg("ocrprep-inverted", invertedImg)

	greyImg := effect.Grayscale(invertedImg)
	writeIntermediateImg("ocrprep-greyscale", greyImg)

	return greyImg
}

// Ignores alpha channel.
func dominantColor(img *image.RGBA, threshold uint8) color.RGBA {
	r := uint32(0)
	g := uint32(0)
	b := uint32(0)
	n := uint32(0)

	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.RGBAAt(x, y)
			if c.R > threshold || c.G > threshold || c.B > threshold {
				r += uint32(c.R)
				g += uint32(c.G)
				b += uint32(c.B)
				n++
			}
		}
	}
	r /= n
	g /= n
	b /= n

	// We don't need to clamp the uint32 -> uint8 conversions as they'll
	// always be less than 256 after dividing by n
	return color.RGBA{uint8(r), uint8(g), uint8(b), 0xff}
}

func delta(a, b uint8) uint32 {
	if a > b {
		return uint32(a - b)
	} else {
		return uint32(b - a)
	}
}

func colorDelta(c1 color.RGBA, c2 color.RGBA) uint32 {
	return delta(c1.R, c2.R) + delta(c1.G, c2.G) + delta(c1.B, c2.B)
}

func wordType(img *image.RGBA) model.KeyphraseType {
	c := dominantColor(img, 0xc0)
	deltaThreshold := uint32(20)
	switch {
	case colorDelta(c, blueColor) < deltaThreshold:
		return model.KeyphraseTypeBlue
	case colorDelta(c, greenColor) < deltaThreshold:
		return model.KeyphraseTypeGreen
	default:
		return model.KeyphraseTypeNone
	}

}

func ocr(img image.Image) (*model.Record, error) {

	contentImg := msxContent(img)
	ocrImg := ocrPrep(contentImg)

	var b bytes.Buffer
	err := png.Encode(&b, ocrImg)
	if err != nil {
		return nil, fmt.Errorf("Failed to endode image to png buffer: %v ", err)
	}

	client := gosseract.NewClient()
	defer client.Close()
	client.SetImageFromBytes(b.Bytes())
	text, err := client.Text()
	if err != nil {
		return nil, err
	}
	writeIntermediateText("ocr", text)

	boxes, err := client.GetBoundingBoxes(gosseract.RIL_WORD)
	if err != nil {
		return nil, err
	}

	normalizedImg := asRGBA(contentImg)
	prevType := model.KeyphraseTypeNone
	keyphrases := map[model.KeyphraseType][]string{}
	for i, box := range boxes {
		wordImg := transform.Crop(normalizedImg, box.Box)
		writeIntermediateImg(fmt.Sprintf("ocr-word-%d-%s", i, box.Word), wordImg)
		wordType := wordType(wordImg)

		trimmedWord := strings.TrimRight(box.Word, ".")

		if wordType != model.KeyphraseTypeNone {
			if wordType == prevType {
				prevIndex := len(keyphrases[wordType]) - 1
				keyphrases[wordType][prevIndex] = keyphrases[wordType][prevIndex] + " " + trimmedWord
			} else {
				keyphrases[wordType] = append(keyphrases[wordType], trimmedWord)
			}
		}

		if trimmedWord != box.Word {
			// Don't aggregate words across sentences.
			prevType = model.KeyphraseTypeNone
		} else {
			prevType = wordType
		}
	}
	record := &model.Record{
		Text:       text,
		Keyphrases: keyphrases,
	}
	writeIntermediateJson("ocr-record", record)

	return record, nil
}
