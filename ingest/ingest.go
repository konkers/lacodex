package ingest

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"

	"github.com/anthonynsimon/bild/effect"
	"github.com/anthonynsimon/bild/transform"
	"github.com/konkers/lacodex/model"
	"github.com/otiai10/gosseract"
)

const nativeWidth = 640
const nativeHeight = 480

const msxContentWidth = 604
const msxContentHeight = 412

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

// UtilCropGameImage is a function to be used by debugging utilities to
// produce and image cropped for the game's size.
func UtilCropGameImage(img image.Image) image.Image {
	return cropGameImage(img)
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

// image Compare compares two images
//
// Returns: likeness factor between 0.0 and 1.0.
//
// If any pixel is not fully opaque (alpha of 0xff) in either image, that pixel
// is not compared.  The comparison assumes that the two images are of the same
// size.
func imageCompare(imgA image.Image, imgB image.Image) float64 {
	a, ok := imgA.(*image.RGBA)
	if !ok {
		a = asRGBA(imgA)
	}

	b, ok := imgB.(*image.RGBA)
	if !ok {
		b = asRGBA(imgB)
	}

	w, h := a.Bounds().Dx(), a.Bounds().Dy()
	aX, aY := a.Bounds().Min.X, a.Bounds().Min.Y
	bX, bY := b.Bounds().Min.X, b.Bounds().Min.Y

	delta := uint32(0)
	n := uint32(0)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cA := a.RGBAAt(aX+x, aY+y)
			cB := b.RGBAAt(bX+x, bY+y)
			if cA.A == 0xff && cB.A == 0xff {
				n++
				delta += colorDelta(cA, cB)
			}
		}
	}

	return 1.0 - float64(delta)/float64(n*3*0xff)
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

func getKeyphrases(client *gosseract.Client, img image.Image) (map[model.KeyphraseType][]string, error) {
	boxes, err := client.GetBoundingBoxes(gosseract.RIL_WORD)
	if err != nil {
		return nil, err
	}

	normalizedImg := asRGBA(img)
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
	text, err := client.Text()
	if err != nil {
		return nil, err
	}
	writeIntermediateText(tag, text)

	keyphrases, err := getKeyphrases(client, img)
	if err != nil {
		return nil, err
	}

	record := &model.Record{
		Text:       text,
		Keyphrases: keyphrases,
	}
	writeIntermediateJson(tag+"-record", record)

	return record, nil
}

func ocrScanner(img image.Image) (*model.Record, error) {
	contentImg := msxContent(img)
	record, err := ocrImage("ocr", contentImg, true)
	if err != nil {
		return nil, err
	}

	record.Type = model.RecordTypeScanner
	return record, nil
}

func ocrTent(img image.Image) (*model.Record, error) {
	imgBounds := img.Bounds()
	cropBounds := image.Rect(105, 125, 535, 310)
	cropBounds.Min = cropBounds.Min.Add(imgBounds.Min)
	cropBounds.Max = cropBounds.Max.Add(imgBounds.Min)

	contentImg := transform.Crop(img, cropBounds)
	record, err := ocrImage("ocr", contentImg, true)
	if err != nil {
		return nil, err
	}

	record.Type = model.RecordTypeTent

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

		c := imageCompare(img, refImg)
		if c > confidence {
			confidence = c
			recordType = t
		}
	}
	return recordType, confidence, nil
}

func IngestImage(img image.Image) (*model.Record, error) {
	img = cropGameImage(img)
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
	default:
		return nil, fmt.Errorf("Can't handle record type %d", recordType)
	}
}
