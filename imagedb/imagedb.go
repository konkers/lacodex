package imagedb

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/png"
	_ "image/png" // Pull in png decoder.
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/konkers/lacodex/model"

	"github.com/asdine/storm"
)

const imagesBucket = "__images__"

// Steam names screenshots like 230700_20190517183348_1.png
var screenshotNameRegexp = regexp.MustCompile(
	`^\d+_(?P<year>\d{4})(?P<month>\d{2})(?P<day>\d{2})` +
		`(?P<hour>\d{2})(?P<minute>\d{2})(?P<second>\d{2})_\d+\.png`)

type ImageDB struct {
	db storm.Node
}

func NewImageDB(db storm.Node) *ImageDB {
	return &ImageDB{db: db}
}

func getScreenshotTime(fileName string) (time.Time, error) {
	m := screenshotNameRegexp.FindStringSubmatch(fileName)
	if m == nil {
		return time.Time{}, fmt.Errorf("%s is not a properly formatted steam screenshot name", fileName)
	}

	// We know these will succeed as they came from the regexp.
	year, _ := strconv.Atoi(m[1])
	month, _ := strconv.Atoi(m[2])
	day, _ := strconv.Atoi(m[3])
	hour, _ := strconv.Atoi(m[4])
	minute, _ := strconv.Atoi(m[5])
	sec, _ := strconv.Atoi(m[6])

	return time.Date(year, time.Month(month), day, hour, minute, sec, 0, time.Local), nil
}

func calcImageHash(img *image.RGBA) string {
	hash := sha256.New()
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		offset := img.PixOffset(b.Min.X, y)
		hash.Write(img.Pix[offset : offset+b.Dx()*4])
	}

	return "sha256-" + hex.EncodeToString(hash.Sum(nil))
}

func encodeImage(img *image.RGBA) ([]byte, error) {
	var buf bytes.Buffer

	err := png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (idb *ImageDB) ImportScreenshot(fileName string, recordId int, img *image.RGBA) error {
	bounds := img.Bounds()
	if bounds.Dx() != 640 && bounds.Dy() != 480 {
		return fmt.Errorf("Image size (%dx%d) was not the expected 640x480", bounds.Dx(), bounds.Dy())
	}

	baseName := filepath.Base(fileName)
	capturedAt, err := getScreenshotTime(baseName)
	if err != nil {
		return err
	}

	hash := calcImageHash(img)
	exists, _ := idb.db.KeyExists(imagesBucket, hash)

	if !exists {
		imgData, err := encodeImage(img)
		if err != nil {
			return err
		}

		err = idb.db.SetBytes(imagesBucket, hash, imgData)
		if err != nil {
			return err
		}
	}

	meta := model.ImageMetadata{
		Hash:       hash,
		CapturedAt: capturedAt,
		FileName:   baseName,
		Record:     recordId,
	}

	return idb.db.Save(&meta)
}

func (idb *ImageDB) LookupFile(fileName string) (*model.ImageMetadata, error) {
	var meta model.ImageMetadata
	err := idb.db.One("FileName", fileName, &meta)
	if err != nil {
		return nil, err
	}

	return &meta, nil
}

func (idb *ImageDB) GetImageData(hash string) ([]byte, error) {
	return idb.db.GetBytes(imagesBucket, hash)
}

func (idb *ImageDB) GetImage(hash string) (image.Image, error) {
	data, err := idb.GetImageData(hash)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (idb *ImageDB) ListImages() ([]*model.ImageMetadata, error) {
	var meta []*model.ImageMetadata
	err := idb.db.All(&meta)
	if err != nil {
		return nil, err
	}
	return meta, nil
}
