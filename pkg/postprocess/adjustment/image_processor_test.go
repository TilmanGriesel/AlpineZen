// SPDX-FileCopyrightText: 2024 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later

package adjustment

import (
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdjustBrightness(t *testing.T) {
	tests := []struct {
		c          float64
		brightness float64
		expected   float64
	}{
		{0.5, 0.2, 0.7},
		{0.0, 0.5, 0.5},
		{1.0, -0.5, 0.5},
		{0.3, -0.5, 0.0},
		{0.8, 0.3, 1.0},
	}

	for _, tt := range tests {
		result := AdjustBrightness(tt.c, tt.brightness)
		assert.Equal(t, tt.expected, result, "adjustBrightness result should match expected value")
	}
}

func TestRotateHue(t *testing.T) {
	r, g, b := 0.5, 0.5, 0.5
	angle := 90.0
	newR, newG, newB := RotateHue(r, g, b, angle)

	assert.InDelta(t, 0.5, newR, 0.1, "newR should be close to 0.5")
	assert.InDelta(t, 0.5, newG, 0.1, "newG should be close to 0.5")
	assert.InDelta(t, 0.5, newB, 0.1, "newB should be close to 0.5")
}

func TestApplyGamma(t *testing.T) {
	tests := []struct {
		c      float64
		gamma  float64
		result float64
	}{
		{0.5, 1.0, 0.5},
		{0.5, 2.0, 0.70710678118},
		{0.5, 0.5, 0.25},
	}

	for _, tt := range tests {
		result := ApplyGamma(tt.c, tt.gamma)
		assert.InDelta(t, tt.result, result, 0.0001, "applyGamma result should match expected value")
	}
}

func TestAdjustBlackPoint(t *testing.T) {
	tests := []struct {
		c          float64
		blackPoint float64
		expected   float64
	}{
		{0.5, 0.2, 0.375},
		{0.2, 0.2, 0.0},
		{0.8, 0.2, 0.75},
		{0.0, 0.1, 0.0},
		{1.0, 0.1, 1.0},
	}

	for _, tt := range tests {
		result := AdjustBlackPoint(tt.c, tt.blackPoint)
		assert.InDelta(t, tt.expected, result, 0.0001, "adjustBlackPoint result should match expected value")
	}
}

func TestAdjustWhitePoint(t *testing.T) {
	tests := []struct {
		c          float64
		whitePoint float64
		expected   float64
	}{
		{0.5, 0.8, 0.625},
		{0.2, 1.0, 0.2},
		{0.8, 0.8, 1.0},
		{0.0, 0.9, 0.0},
	}

	for _, tt := range tests {
		result := AdjustWhitePoint(tt.c, tt.whitePoint)
		assert.InDelta(t, tt.expected, result, 0.0001, "adjustWhitePoint result should match expected value")
	}
}

func TestAdjustShadowStrength(t *testing.T) {
	tests := []struct {
		c        float64
		strength float64
		expected float64
	}{
		{0.5, 2.0, 1.0},
		{0.2, 0.5, 0.1},
		{0.8, 1.5, 1.0},
		{0.0, 1.0, 0.0},
		{1.0, 1.0, 1.0},
	}

	for _, tt := range tests {
		result := AdjustShadowStrength(tt.c, tt.strength)
		assert.Equal(t, tt.expected, result, "adjustShadowStrength result should match expected value")
	}
}

func TestImageEnhancerApplyEnhancements(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	enhancer := NewImageEnhancer()
	enhancedImg := enhancer.ApplyEnhancements(img)

	assert.NotNil(t, enhancedImg, "ApplyEnhancements should return a non-nil image")
	assert.Equal(t, img.Bounds(), enhancedImg.Bounds(), "Enhanced image should have same dimensions as original image")
}
