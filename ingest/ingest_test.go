package ingest

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	_ "image/png" // Pull in png decoder.
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/konkers/lacodex/model"

	"github.com/anthonynsimon/bild/util"
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
		ocrText: `There are 8 Ankhs.

8 Ankhs that protect the great spirits.
	
Seek the red light; the Ankh Jewel.
	
The guardians that slumber within the Ankh will test
thine strength.

OK

f`,
	},
	testImageDesc{
		name: "screenshot2",
		ocrText: `"The first age of the sun was destroyed by flood,

		The second age of the sun was destroyed by the god of wind,
		The third age of the sun was destroyed by the god of fire,
		The fourth age of the sun was destroyed by blood and fire
		falling from the sky."
		
		The same thing was written in Mayan prophecy.
		
		Could there be a connection?`,
	},
}

func imageCompare(t *testing.T, name string, tag string, testImg image.Image, goldImg image.Image) bool {
	return util.RGBAImageEqual(asRGBA(testImg), asRGBA(goldImg))
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

	if !imageCompare(t, name, tag, img, goldImg) {
		t.Errorf("%s-%s differs from the gold image.", name, tag)
	}
}

func TestPrepImage(t *testing.T) {
	for _, i := range testImages {

		imgFile := fmt.Sprintf("test_data/%s.png", i.name)
		reader, err := os.Open(imgFile)
		if err != nil {
			t.Fatal(err)
		}
		defer reader.Close()

		img, _, err := image.Decode(reader)
		if err != nil {
			t.Fatal(err)
		}

		if writeIntermediates {
			intermediatePrefix = i.name
		}

		gameImg := cropGameImage(img)
		testImage(t, i.name, "testout-game", gameImg)

		contentImg := msxContent(gameImg)
		testImage(t, i.name, "testout-content", contentImg)

		ocrImg := ocrPrep(contentImg)
		testImage(t, i.name, "testout-ocrprep", ocrImg)

		record, err := ocr(gameImg)
		if err != nil {
			t.Error(err)
			continue
		}

		b, err := ioutil.ReadFile(fmt.Sprintf("test_data/%s-ocr-record.json", i.name))
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

		if !reflect.DeepEqual(record, &goldRecord) {
			t.Errorf("%s record does not match gold", i.name)
			continue
		}

	}
}
