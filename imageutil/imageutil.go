package imageutil

import (
	"image"
	"image/color"
	"image/draw"
)

func AsRGBA(src image.Image) *image.RGBA {
	srcBounds := src.Bounds()
	destBounds := image.Rect(0, 0, srcBounds.Dx(), srcBounds.Dy())
	img := image.NewRGBA(destBounds)
	draw.Draw(img, destBounds, src, srcBounds.Min, draw.Src)
	return img
}

func delta(a, b uint8) uint32 {
	if a > b {
		return uint32(a - b)
	} else {
		return uint32(b - a)
	}
}

func ColorDelta(c1 color.RGBA, c2 color.RGBA) uint32 {
	return delta(c1.R, c2.R) + delta(c1.G, c2.G) + delta(c1.B, c2.B)
}

// ImageCompare compares two images
//
// Returns: likeness factor between 0.0 and 1.0.
//
// If any pixel is not fully opaque (alpha of 0xff) in either image, that pixel
// is not compared.  The comparison assumes that the two images are of the same
// size.
func ImageCompare(imgA image.Image, imgB image.Image) float64 {
	a, ok := imgA.(*image.RGBA)
	if !ok {
		a = AsRGBA(imgA)
	}

	b, ok := imgB.(*image.RGBA)
	if !ok {
		b = AsRGBA(imgB)
	}

	w, h := a.Bounds().Dx(), a.Bounds().Dy()
	aX, aY := a.Bounds().Min.X, a.Bounds().Min.Y
	bX, bY := b.Bounds().Min.X, b.Bounds().Min.Y

	delta := uint32(0)
	n := uint32(0)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cA := a.RGBAAt(aX+x, aY+y)
			cB := b.RGBAAt(bX+x, bY+y)
			if cA.A == 0xff && cB.A == 0xff {
				n++
				delta += ColorDelta(cA, cB)
			}
		}
	}

	return 1.0 - float64(delta)/float64(n*3*0xff)
}
