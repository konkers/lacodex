package testutil

import (
	"fmt"
	"image"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/konkers/lacodex/imageutil"
)

var testRNG = rand.New(rand.NewSource(928084234))

func writeImage(t *testing.T, fileName string, img image.Image) {
	os.MkdirAll("testout", 0755)

	writer, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("Failed to test output image %s: %v\n", fileName, err)
		return
	}

	err = png.Encode(writer, img)
	if err != nil {
		fmt.Printf("Failed to endode test output timage %s: %v\n", fileName, err)
		return
	}

}

func AssertImagesEqual(t *testing.T, expectedImg image.Image, actualImg image.Image) {
	likeness := imageutil.ImageCompare(expectedImg, actualImg)
	if likeness < 1.0 {

		num := testRNG.Int31n(999999999)
		expectedName := filepath.Join("testout", fmt.Sprintf("%d-expected.png", num))
		actualName := filepath.Join("testout", fmt.Sprintf("%d-actual.png", num))

		writeImage(t, expectedName, expectedImg)
		writeImage(t, actualName, actualImg)

		t.Fatalf("Images are only %f equal.  Look at %s and %s.",
			likeness, expectedName, actualName)
	}
}

func LoadTestImage(t *testing.T, imgFile string) image.Image {
	reader, err := os.Open(imgFile)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	img, _, err := image.Decode(reader)
	if err != nil {
		t.Fatal(err)
	}

	return img
}
