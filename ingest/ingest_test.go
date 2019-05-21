package ingest

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	_ "image/png" // Pull in png decoder.
	"io/ioutil"
	"os"
	"testing"

	"github.com/anthonynsimon/bild/util"
	"github.com/konkers/lacodex/imageutil"
	"github.com/konkers/lacodex/model"
	"github.com/konkers/lacodex/testutil"
	"github.com/stretchr/testify/assert"
)

var writeIntermediates bool

func init() {
	flag.BoolVar(&writeIntermediates, "write-intermediates", false, "Write intermediates?")
}

type testImageDesc struct {
	name    string
	ocrText string
}

var testImages = []testImageDesc{
	testImageDesc{
		name: "screenshot1",
	},
	testImageDesc{
		name: "screenshot2",
	},
}

func testImagesEqual(t *testing.T, name string, tag string, testImg image.Image, goldImg image.Image) bool {
	return util.RGBAImageEqual(imageutil.AsRGBA(testImg), imageutil.AsRGBA(goldImg))
}

func testImage(t *testing.T, name string, tag string, img image.Image) {
	if writeIntermediates {
		writeIntermediateImg(tag, img)
	}
	goldImgFile := fmt.Sprintf("test_data/%s-%s.png", name, tag)
	reader, err := os.Open(goldImgFile)
	if err != nil {
		t.Error(err)
		return
	}
	defer reader.Close()

	goldImg, _, err := image.Decode(reader)
	if err != nil {
		t.Error(err)
		return
	}

	if !testImagesEqual(t, name, tag, img, goldImg) {
		t.Errorf("%s-%s differs from the gold image.", name, tag)
	}
}

func loadTestImage(t *testing.T, name string) image.Image {
	return testutil.LoadTestImage(t, fmt.Sprintf("test_data/%s.png", name))
}

func TestGameCropImage(t *testing.T) {
	testImages := []string{"screenshot1", "screenshot2"}
	for _, name := range testImages {
		img := loadTestImage(t, name)

		if writeIntermediates {
			intermediatePrefix = "testGameCrop-" + name
		}

		gameImg := CropGameImage(img)
		testImage(t, name, "testout-game", gameImg)
	}
}
func TestClassifyImage(t *testing.T) {
	tests := []struct {
		name string
		t    model.RecordType
	}{
		{"classify-tent0", model.RecordTypeTent},
		{"classify-tent1", model.RecordTypeTent},
		{"classify-mailer0", model.RecordTypeMailer},
		{"classify-mailer1", model.RecordTypeMailer},
		{"screenshot1", model.RecordTypeScanner},
		{"screenshot2", model.RecordTypeScanner},
	}

	for _, test := range tests {
		img := loadTestImage(t, test.name)
		gameImg := CropGameImage(img)
		recordType, confidence, err := classifyImage(gameImg)
		if err != nil {
			t.Errorf("Failed to classify %s: %v", test.name, err)
			continue
		}

		assert.Equal(t, recordType, test.t)
		assert.GreaterOrEqual(t, confidence, 0.9)
	}
}

func TestIngest(t *testing.T) {
	tests := []string{
		"classify-tent0",
		"classify-tent1",
		"screenshot1",
		"screenshot2",
		"classify-mailer0",
		"classify-mailer1",
	}
	for _, name := range tests {
		if writeIntermediates {
			intermediatePrefix = "testIngest-" + name
		}
		img := loadTestImage(t, name)
		gameImg := CropGameImage(img)
		record, err := IngestImage(gameImg)
		if err != nil {
			t.Errorf("Failed to classify %s: %v", name, err)
			return
		}

		b, err := ioutil.ReadFile(fmt.Sprintf("test_data/%s-record.json", name))
		if err != nil {
			t.Error(err)
			continue
		}
		var goldRecord model.Record
		err = json.Unmarshal(b, &goldRecord)
		if err != nil {
			t.Error(err)
			continue
		}

		assert.Equal(t, &goldRecord, record)
	}
}
