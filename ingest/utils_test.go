package ingest

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/konkers/lacodex/testutil"

	"github.com/stretchr/testify/assert"
)

func TestWriteIntermediateText(t *testing.T) {

	fileName := "./intermediates/utils-test-text-1.txt"
	intermediatePrefix = "utils-test-text"
	writeIntermediateText("1", "test")
	defer os.Remove(fileName)

	buf, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "test", string(buf))
}

func TestWriteIntermediateJson(t *testing.T) {
	fileName := "./intermediates/utils-test-json-1.json"
	intermediatePrefix = "utils-test-json"
	writeIntermediateJson("1", "test")
	defer os.Remove(fileName)

	buf, err := ioutil.ReadFile(fileName)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "\"test\"", string(buf))
}

func TestWriteIntermediateImg(t *testing.T) {
	img := loadTestImage(t, "screenshot1")

	fileName := "./intermediates/utils-test-img-1.png"
	intermediatePrefix = "utils-test-img"
	writeIntermediateImg("1", img)
	defer os.Remove(fileName)

	interImg := testutil.LoadTestImage(t, fileName)
	testutil.AssertImagesEqual(t, img, interImg)
}
