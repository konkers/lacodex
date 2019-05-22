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

type testIdb struct {
	Idb      *ImageDB
	Filename string
	Db       *storm.DB
}

func newTestImageDB(t *testing.T) *testIdb {
	tmpfile, err := ioutil.TempFile("", "imagedbtest.*.db")
	if err != nil {
		t.Fatal(err)
	}

	db, err := storm.Open(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	bucket := db.From("imagedb")
	idb := NewImageDB(bucket)

	return &testIdb{
		Idb:      idb,
		Filename: tmpfile.Name(),
		Db:       db,
	}
}

func (i *testIdb) Close() {
	i.Db.Close()
	os.Remove(i.Filename)
}
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

	hashA := calcImageHash(imgA)
	hashB := calcImageHash(imgB)

	assert.Equal(t, hashA, hashB)
}

func TestCalcImageHashDiff(t *testing.T) {
	imgA := ingest.CropGameImage(testutil.LoadTestImage(t, "../testdata/screenshots/230700_20190517185334_1.png"))
	imgB := ingest.CropGameImage(testutil.LoadTestImage(t, "../testdata/screenshots/230700_20190519134145_1.png"))

	hashA := calcImageHash(imgA)
	hashB := calcImageHash(imgB)

	assert.NotEqual(t, hashA, hashB)
}

func TestEncodeImageFailure(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 0, 0))
	_, err := encodeImage(img)
	if err == nil {
		t.Error("Expexcted error")
	}
}

func TestListImagesFailure(t *testing.T) {
	testIdb := newTestImageDB(t)
	idb := testIdb.Idb
	testIdb.Close()

	_, err := idb.ListImages()
	if err == nil {
		t.Error("Expected error")
	}
}

func TestImportScreenshot(t *testing.T) {
	testIdb := newTestImageDB(t)
	defer testIdb.Close()
	idb := testIdb.Idb

	imgA := ingest.CropGameImage(testutil.LoadTestImage(t, "../testdata/screenshots/230700_20190519134140_1.png"))
	imgB := ingest.CropGameImage(testutil.LoadTestImage(t, "../testdata/screenshots/230700_20190519134145_1.png"))

	err := idb.ImportScreenshot("230700_20190519134140_1.png", 1, imgA)
	if err != nil {
		t.Fatal(err)
	}
	err = idb.ImportScreenshot("230700_20190519134145_1.png", 2, imgB)
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
		Id:         1,
		Hash:       "sha256-8acc37faaea0c3ff4ea847288a76e5f713d5857f14bf6dbdeffe7dbedd3234db",
		CapturedAt: time.Date(2019, time.Month(5), 19, 13, 41, 40, 0, time.Local),
		FileName:   "230700_20190519134140_1.png",
		Record:     1,
	}, metaA)

	assert.Equal(t, &model.ImageMetadata{
		Id:         2,
		Hash:       "sha256-8acc37faaea0c3ff4ea847288a76e5f713d5857f14bf6dbdeffe7dbedd3234db",
		CapturedAt: time.Date(2019, time.Month(5), 19, 13, 41, 45, 0, time.Local),
		FileName:   "230700_20190519134145_1.png",
		Record:     2,
	}, metaB)

	img, err := idb.GetImage(metaA.Hash)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertImagesEqual(t, imgA, img)

	meta, err := idb.ListImages()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []*model.ImageMetadata{metaA, metaB}, meta)

	// Test failure case: Unencodable image.
	img = image.NewRGBA(image.Rect(0, 0, 0, 0))
	err = idb.ImportScreenshot("230700_20190519134145_1.png", 1, img.(*image.RGBA))
	if err == nil {
		t.Fatal("Expected error")
	}

	// Test failure case: Bad file name.
	err = idb.ImportScreenshot("230700_2019051913414_1.png", 1, imgA)
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
