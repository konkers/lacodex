package ingest

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestGetReferenceImageFail(t *testing.T) {
	badFile := "reference/bad.png"
	err := ioutil.WriteFile(badFile, []byte{0}, 0644)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(badFile)

	_, err = getReferenceImage("kdfjlskjfasdf")
	if err == nil {
		t.Error("Expected Error")
	}

	_, err = getReferenceImage("bad")
	if err == nil {
		t.Error("Expected Error")
	}
}
