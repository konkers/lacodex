package ingest

import (
	"bytes"
	"image"
	"io/ioutil"
	"log"
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

	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stdout)

	// Un-writable intermediate
	os.Chmod(fileName, 0)
	writeIntermediateText("1", "test")
	if logBuf.Len() == 0 {
		t.Error("Expected warning to have been logged")
	}
	logBuf.Reset()
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

	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stdout)

	// Un-encodable JSON
	writeIntermediateJson("1", make(chan int))
	if logBuf.Len() == 0 {
		t.Error("Expected warning to have been logged")
	}
	logBuf.Reset()

	// Un-writable intermediate
	os.Chmod(fileName, 0)
	writeIntermediateJson("1", "test")
	if logBuf.Len() == 0 {
		t.Error("Expected warning to have been logged")
	}
	logBuf.Reset()
}

func TestWriteIntermediateImg(t *testing.T) {
	img := loadTestImage(t, "screenshot1")

	fileName := "./intermediates/utils-test-img-1.png"
	intermediatePrefix = "utils-test-img"
	writeIntermediateImg("1", img)
	defer os.Remove(fileName)

	interImg := testutil.LoadTestImage(t, fileName)
	testutil.AssertImagesEqual(t, img, interImg)

	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stdout)

	// Un-encodable image.
	img = image.NewRGBA(image.Rect(0, 0, 0, 0))
	writeIntermediateImg("1", img)
	if logBuf.Len() == 0 {
		t.Error("Expected warning to have been logged")
	}
	logBuf.Reset()

	// Un-writable intermediate
	os.Chmod(fileName, 0)
	writeIntermediateImg("1", img)
	if logBuf.Len() == 0 {
		t.Error("Expected warning to have been logged")
	}
	logBuf.Reset()
}
