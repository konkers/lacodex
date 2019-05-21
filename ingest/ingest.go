package ingest

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strconv"
	"strings"

	"github.com/anthonynsimon/bild/effect"
	"github.com/anthonynsimon/bild/transform"
	"github.com/konkers/lacodex/imageutil"
	"github.com/konkers/lacodex/model"
	"github.com/otiai10/gosseract"
)

const nativeWidth = 640
const nativeHeight = 480

const msxContentWidth = 604
const msxContentHeight = 412

const confidenceThreshold = 60

var normalColor = color.RGBA{230, 232, 236, 255}
var blueColor = color.RGBA{100, 182, 227, 255}
var greenColor = color.RGBA{96, 229, 147, 255}

func middleCrop(img image.Image, width int, height int) image.Image {
	bounds := img.Bounds()
	insetX := bounds.Min.X + (bounds.Dx()-width)/2
	insetY := bounds.Min.Y + (bounds.Dy()-height)/2
	cropRect := image.Rect(insetX, insetY, width+insetX, height+insetY)
	return transform.Crop(img, cropRect)
}

func CropGameImage(img image.Image) image.Image {
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
func ocrPrep(img image.Image, invert bool) image.Image {
	if invert {
		img = effect.Invert(img)
		writeIntermediateImg("ocrprep-inverted", img)
	}

	greyImg := effect.Grayscale(img)
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

func wordType(img *image.RGBA) model.KeyphraseType {
	c := dominantColor(img, 0xc0)
	deltaThreshold := uint32(20)
	switch {
	case imageutil.ColorDelta(c, blueColor) < deltaThreshold:
		return model.KeyphraseTypeBlue
	case imageutil.ColorDelta(c, greenColor) < deltaThreshold:
		return model.KeyphraseTypeGreen
	default:
		return model.KeyphraseTypeNone
	}

}

func getKeyphrases(client *gosseract.Client, img image.Image) (map[model.KeyphraseType][]string, error) {
	boxes, err := client.GetBoundingBoxes(gosseract.RIL_WORD)
	if err != nil {
		return nil, err
	}

	normalizedImg := imageutil.AsRGBA(img)
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
	return keyphrases, nil
}

func ocrImage(tag string, img image.Image, invert bool) (*model.Record, error) {
	ocrImg := ocrPrep(img, invert)

	// There should be some better way to pass this image into tesseract, but
	// I can't find one.
	var b bytes.Buffer
	err := png.Encode(&b, ocrImg)
	if err != nil {
		return nil, fmt.Errorf("Failed to endode image to png buffer: %v ", err)
	}

	client := gosseract.NewClient()
	defer client.Close()
	client.SetImageFromBytes(b.Bytes())

	text := ""
	boxes, err := client.GetBoundingBoxes(gosseract.RIL_PARA)
	if err != nil {
		return nil, err
	}
	for _, box := range boxes {
		if box.Confidence > confidenceThreshold {
			text += box.Word
		}
	}
	text = strings.TrimSpace(text)
	writeIntermediateText(tag, text)

	if text == "" {
		return nil, fmt.Errorf("No text found in image")
	}

	if text == "OK" {
		return nil, fmt.Errorf("Image is untranslated glyphs")
	}

	keyphrases, err := getKeyphrases(client, img)
	if err != nil {
		return nil, err
	}

	record := &model.Record{
		Text:       text,
		Keyphrases: keyphrases,
	}
	return record, nil
}

func ocrTextAt(tag string, img image.Image, rect image.Rectangle, invert bool) (*model.Record, error) {
	bounds := imageutil.OffsetRect(rect, img.Bounds())
	return ocrImage(tag, transform.Crop(img, bounds), invert)
}

func ocrNumbersAt(tag string, img image.Image, rect image.Rectangle, invert bool) (*model.Record, error) {
	bounds := imageutil.OffsetRect(rect, img.Bounds())
	record, err := ocrImage(tag, transform.Crop(img, bounds), invert)
	if err != nil {
		return nil, err
	}
	record.Text = strings.Map(func(r rune) rune {
		switch r {
		case 'o':
			return '0'
		case 'l':
			return '1'
		default:
			return r
		}
	}, record.Text)

	return record, nil
}

func ocrScanner(img image.Image) (*model.Record, error) {
	contentImg := msxContent(img)
	record, err := ocrImage("ocr", contentImg, true)
	if err != nil {
		return nil, err
	}
	record.Type = model.RecordTypeScanner

	writeIntermediateJson("record", record)
	return record, nil
}

func ocrTent(img image.Image) (*model.Record, error) {
	record, err := ocrTextAt("ocr", img, image.Rect(105, 125, 535, 310), true)
	if err != nil {
		return nil, err
	}
	record.Type = model.RecordTypeTent

	writeIntermediateJson("record", record)
	return record, nil
}

func ocrMailer(img image.Image) (*model.Record, error) {
	record, err := ocrTextAt("ocr", img, image.Rect(18, 178, 622, 446), false)
	if err != nil {
		return nil, err
	}

	indexRecord, err := ocrNumbersAt("ocr-index", img, image.Rect(47, 74, 73, 91), false)
	if err != nil {
		return nil, err
	}
	index, err := strconv.Atoi(indexRecord.Text)
	if err != nil {
		return nil, err
	}
	record.Index = &index

	subjectRecord, err := ocrTextAt("ocr", img, image.Rect(77, 74, 523, 92), false)
	if err != nil {
		return nil, err
	}
	record.Subject = subjectRecord.Text

	record.Type = model.RecordTypeMailer

	writeIntermediateJson("record", record)
	return record, nil
}

// returns a RecordType, confidence tuple.
func classifyImage(img image.Image) (model.RecordType, float64, error) {
	types := []model.RecordType{
		model.RecordTypeTent,
		model.RecordTypeMailer,
		model.RecordTypeScanner,
	}

	var recordType model.RecordType
	var confidence float64
	for _, t := range types {
		nameB, _ := t.MarshalText()
		name := string(nameB)
		refImg, err := getReferenceImage(name)
		if err != nil {
			return model.RecordTypeTent, 0.0, err
		}

		c := imageutil.ImageCompare(img, refImg)
		if c > confidence {
			confidence = c
			recordType = t
		}
	}
	return recordType, confidence, nil
}

func IngestImage(img image.Image) (*model.Record, error) {
	img = CropGameImage(img)
	recordType, confidence, err := classifyImage(img)
	if err != nil {
		return nil, err
	}
	if confidence < 0.9 {
		return nil, fmt.Errorf("Image classification confidence %f is not hight enough.", confidence)
	}

	switch recordType {
	case model.RecordTypeTent:
		return ocrTent(img)
	case model.RecordTypeScanner:
		return ocrScanner(img)
	case model.RecordTypeMailer:
		return ocrMailer(img)
	default:
		return nil, fmt.Errorf("Can't handle record type %d", recordType)
	}
}
