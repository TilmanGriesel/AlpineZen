// SPDX-FileCopyrightText: 2025 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later AND LicenseRef-AlpineZen-Trademark

package postprocess

import (
	"image"
	"image/color"
	"image/draw"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateAverageBrightness(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.NRGBA{0x80, 0x80, 0x80, 0xff}}, image.Point{}, draw.Src)
	avgBrightness := CalculateAverageBrightness(img)
	expectedBrightness := 0.5

	assert.InDelta(t, expectedBrightness, avgBrightness, 0.01, "calculateAverageBrightness should return expected average brightness")
}

func TestCalculateScaledOpacity(t *testing.T) {
	tests := []struct {
		avgBrightness float64
		minBrightness float64
		maxBrightness float64
		minOpacity    float64
		maxOpacity    float64
		expected      float64
	}{
		{0.2, 0.0, 0.4, 0.01, 0.5, 0.255},
		{0.0, 0.0, 0.4, 0.01, 0.5, 0.01},
		{0.4, 0.0, 0.4, 0.01, 0.5, 0.5},
		{0.5, 0.0, 0.4, 0.01, 0.5, 0.5}, // avgBrightness clamped to maxBrightness
	}

	for _, tt := range tests {
		result := CalculateScaledOpacity(tt.avgBrightness, tt.minBrightness, tt.maxBrightness, tt.minOpacity, tt.maxOpacity)
		assert.InDelta(t, tt.expected, result, 0.01, "calculateScaledOpacity should return expected opacity")
	}
}

func TestCreateNoiseImage(t *testing.T) {
	width, height := 100, 100
	noiseImg := CreateNoiseImage(width, height)

	assert.Equal(t, width, noiseImg.Bounds().Dx(), "Noise image should have expected width")
	assert.Equal(t, height, noiseImg.Bounds().Dy(), "Noise image should have expected height")
}
