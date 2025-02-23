// SPDX-FileCopyrightText: 2025 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later AND LicenseRef-AlpineZen-Trademark

package postprocess

import (
	"image"
	"image/color"
)

func CalculateAverageBrightness(img image.Image) float64 {
	var totalBrightness float64
	bounds := img.Bounds()
	pixelCount := bounds.Dx() * bounds.Dy()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			brightness := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
			totalBrightness += brightness / 65535 // Normalize to 0-1 range
		}
	}

	return totalBrightness / float64(pixelCount)
}

func CalculateScaledOpacity(avgBrightness, minBrightness, maxBrightness, minOpacity, maxOpacity float64) float64 {
	if avgBrightness < minBrightness {
		avgBrightness = minBrightness
	} else if avgBrightness > maxBrightness {
		avgBrightness = maxBrightness
	}

	scaledOpacity := minOpacity + (maxOpacity-minOpacity)*((avgBrightness-minBrightness)/(maxBrightness-minBrightness))

	if scaledOpacity < 0.0 {
		scaledOpacity = 0.0
	} else if scaledOpacity > 1.0 {
		scaledOpacity = 1.0
	}

	return scaledOpacity
}

func CreateNoiseImage(width, height int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var v uint8
			if getRandomBoolWithWrap() {
				v = 255
			} else {
				v = 0
			}
			img.SetNRGBA(x, y, color.NRGBA{v, v, v, 255})
		}
	}
	return img
}
