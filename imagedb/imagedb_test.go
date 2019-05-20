package imagedb

import (
	"image"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/konkers/lacodex/model"
	"github.com/stretchr/testify/assert"

	"github.com/asdine/storm"
	"github.com/konkers/lacodex/ingest"
	"github.com/konkers/lacodex/testutil"
)

func TestGetScreenshotTime(t *testing.T) {
	timestamp, err := getScreenshotTime("230700_20190517183348_1.png")
	if err != nil {
		t.Fatal(err)
	}

	expected := time.Date(2019, time.Month(5), 17, 18, 33, 48, 0, time.Local)
	if !timestamp.Equal(expected) {
		t.Fatalf("Got %v, expected %v", timestamp, expected)
	}

	_, err = getScreenshotTime("230700_2019051718334_1.png")
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestCalcImageHash(t *testing.T) {
	imgA := ingest.CropGameImage(testutil.LoadTestImage(t, "../testdata/screenshots/230700_20190519134140_1.png"))
	imgB := ingest.CropGameImage(testutil.LoadTestImage(t, "../testdata/screenshots/230700_20190519134145_1.png"))

	hashA := calcImageHash(imgA.(*image.RGBA))
	hashB := calcImageHash(imgB.(*image.RGBA))

	assert.Equal(t, hashA, hashB)
}

func TestEncodeImageFailure(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 0, 0))
	_, err := encodeImage(img)
	if err == nil {
		t.Error("Expexcted error")
	}
}
func TestImportScreenshot(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "imagedbtest.*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	db, err := storm.Open(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	bucket := db.From("imagedb")
	idb := NewImageDB(bucket)

	imgA := ingest.CropGameImage(testutil.LoadTestImage(t, "../testdata/screenshots/230700_20190519134140_1.png"))
	imgB := ingest.CropGameImage(testutil.LoadTestImage(t, "../testdata/screenshots/230700_20190519134145_1.png"))

	err = idb.ImportScreenshot("230700_20190519134140_1.png", imgA.(*image.RGBA))
	if err != nil {
		t.Fatal(err)
	}
	err = idb.ImportScreenshot("230700_20190519134145_1.png", imgB.(*image.RGBA))
	if err != nil {
		t.Fatal(err)
	}

	metaA, err := idb.LookupFile("230700_20190519134140_1.png")
	if err != nil {
		t.Fatal(err)
	}

	metaB, err := idb.LookupFile("230700_20190519134145_1.png")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, &model.ImageMetadata{
		Pk:         1,
		Hash:       "sha256-e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		CapturedAt: time.Date(2019, time.Month(5), 19, 13, 41, 40, 0, time.Local),
		FileName:   "230700_20190519134140_1.png",
	}, metaA)

	assert.Equal(t, &model.ImageMetadata{
		Pk:         2,
		Hash:       "sha256-e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		CapturedAt: time.Date(2019, time.Month(5), 19, 13, 41, 45, 0, time.Local),
		FileName:   "230700_20190519134145_1.png",
	}, metaB)

	img, err := idb.GetImage(metaA.Hash)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertImagesEqual(t, imgA, img)

	// Test failure case: Unencodable image.
	img = image.NewRGBA(image.Rect(0, 0, 0, 0))
	err = idb.ImportScreenshot("230700_20190519134145_1.png", img.(*image.RGBA))
	if err == nil {
		t.Fatal("Expected error")
	}

	// Test failure case: Bad file name.
	err = idb.ImportScreenshot("230700_2019051913414_1.png", imgA.(*image.RGBA))
	if err == nil {
		t.Fatal("Expected error")
	}

	// Test failure case: Hash not found.
	_, err = idb.GetImage("")
	if err == nil {
		t.Fatal("Expected error")
	}

	// Test failure case: Bad image data.
	idb.db.SetBytes(imagesBucket, "testhash", []byte{0x0})
	_, err = idb.GetImage("testhash")
	if err == nil {
		t.Fatal("Expected error")
	}
}
