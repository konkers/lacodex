package testutil

import (
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"os"
	"testing"
)

type fakeT struct {
	failed bool
}

func (t *fakeT) Fatal(args ...interface{}) {
	t.failed = true
}

func (t *fakeT) Fatalf(format string, args ...interface{}) {
	t.failed = true
}

func TestAssertImagesEqualFailure(t *testing.T) {
	var ft fakeT

	black := color.RGBA{0, 0, 0, 255}
	white := color.RGBA{255, 255, 255, 255}

	a := image.NewRGBA(image.Rect(0, 0, 100, 100))
	b := image.NewRGBA(image.Rect(0, 0, 100, 100))

	draw.Draw(a, a.Bounds(), &image.Uniform{black}, image.ZP, draw.Src)
	draw.Draw(b, b.Bounds(), &image.Uniform{white}, image.ZP, draw.Src)
	AssertImagesEqual(&ft, a, b)

	if ft.failed == false {
		t.Fatal("Expected failure.")
	}
}

func TestLoadTestImageNoFile(t *testing.T) {
	var ft fakeT

	// This is a bit fragile as it assumes the below files does not exist.
	LoadTestImage(&ft, "dlkfjalksdfjlaksdjflkasdjflkjasdlfjlasdkjf")
	if ft.failed == false {
		t.Fatal("Expected failure.")
	}
}

func TestLoadTestImageBadFile(t *testing.T) {
	var ft fakeT

	f, err := ioutil.TempFile("", "*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	LoadTestImage(&ft, f.Name())
	if ft.failed == false {
		t.Fatal("Expected failure.")
	}
}

func TestWriteImageFailure(t *testing.T) {
	var ft fakeT

	// Un-encodable image.
	img := image.NewRGBA(image.Rect(0, 0, 0, 0))
	writeImage(&ft, "test.png", img)
	if ft.failed == false {
		t.Fatal("Expected failure.")
	}
	ft.failed = false
	defer os.Remove("testout/test.png")

	// Un-openable image.
	os.Chmod("testout/test.png", 0)
	img = image.NewRGBA(image.Rect(0, 0, 100, 100))
	writeImage(&ft, "test.png", img)
	if ft.failed == false {
		t.Fatal("Expected failure.")
	}
}
