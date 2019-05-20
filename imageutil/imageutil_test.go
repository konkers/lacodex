package imageutil

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImageCompare(t *testing.T) {
	clear := color.RGBA{0, 0, 0, 0}
	black := color.RGBA{0, 0, 0, 255}
	white := color.RGBA{255, 255, 255, 255}

	a := image.NewRGBA(image.Rect(0, 0, 100, 100))
	b := image.NewRGBA(image.Rect(0, 0, 100, 100))
	halfBounds := image.Rect(0, 0, 100, 50)

	draw.Draw(a, a.Bounds(), &image.Uniform{black}, image.ZP, draw.Src)
	draw.Draw(b, b.Bounds(), &image.Uniform{white}, image.ZP, draw.Src)
	assert.InDelta(t, 0.0, ImageCompare(a, b), 1e-9)

	draw.Draw(a, a.Bounds(), &image.Uniform{white}, image.ZP, draw.Src)
	draw.Draw(b, b.Bounds(), &image.Uniform{white}, image.ZP, draw.Src)
	assert.InDelta(t, 1.0, ImageCompare(a, b), 1e-9)

	draw.Draw(a, a.Bounds(), &image.Uniform{black}, image.ZP, draw.Src)
	draw.Draw(b, b.Bounds(), &image.Uniform{black}, image.ZP, draw.Src)
	assert.InDelta(t, 1.0, ImageCompare(a, b), 1e-9)

	draw.Draw(a, a.Bounds(), &image.Uniform{black}, image.ZP, draw.Src)
	draw.Draw(b, b.Bounds(), &image.Uniform{black}, image.ZP, draw.Src)
	draw.Draw(b, halfBounds, &image.Uniform{white}, image.ZP, draw.Src)
	assert.InDelta(t, 0.5, ImageCompare(a, b), 1e-9)

	draw.Draw(a, a.Bounds(), &image.Uniform{black}, image.ZP, draw.Src)
	draw.Draw(b, b.Bounds(), &image.Uniform{black}, image.ZP, draw.Src)
	draw.Draw(b, halfBounds, &image.Uniform{white}, image.ZP, draw.Src)
	draw.Draw(a, halfBounds, &image.Uniform{clear}, image.ZP, draw.Src)
	assert.InDelta(t, 1.0, ImageCompare(a, b), 1e-9)
}

func TestImageCompareNonRGBA(t *testing.T) {
	black := color.Gray{0}
	white := color.Gray{255}

	a := image.NewGray(image.Rect(0, 0, 100, 100))
	b := image.NewGray(image.Rect(0, 0, 100, 100))
	draw.Draw(a, a.Bounds(), &image.Uniform{black}, image.ZP, draw.Src)
	draw.Draw(b, b.Bounds(), &image.Uniform{white}, image.ZP, draw.Src)
	assert.InDelta(t, 0.0, ImageCompare(a, b), 1e-9)
}
