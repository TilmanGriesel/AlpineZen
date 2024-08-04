// SPDX-FileCopyrightText: 2024 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later

package adjustment

import (
	"image"
	"image/color"

	"github.com/disintegration/imaging"
)

const (
	NoiseBrightnessThreshold = 0.0
)

type ImageEnhancer struct {
	Contrast       float64
	Saturation     float64
	Brightness     float64
	Hue            float64
	Gamma          float64
	BlackPoint     float64
	WhitePoint     float64
	ShadowStrength float64
}

func NewImageEnhancer() *ImageEnhancer {
	return &ImageEnhancer{
		Contrast:       1.0,
		Saturation:     1.0,
		Brightness:     0.0,
		Hue:            0.0,
		Gamma:          1.0,
		BlackPoint:     0.0,
		WhitePoint:     1.0,
		ShadowStrength: 1.0,
	}
}

func (e *ImageEnhancer) ApplyEnhancements(img image.Image) image.Image {
	return imaging.AdjustFunc(img, func(c color.NRGBA) color.NRGBA {
		r := float64(c.R) / 255.0
		g := float64(c.G) / 255.0
		b := float64(c.B) / 255.0

		r = AdjustBlackPoint(r, e.BlackPoint)
		g = AdjustBlackPoint(g, e.BlackPoint)
		b = AdjustBlackPoint(b, e.BlackPoint)

		r = AdjustWhitePoint(r, e.WhitePoint)
		g = AdjustWhitePoint(g, e.WhitePoint)
		b = AdjustWhitePoint(b, e.WhitePoint)

		r = (r-0.5)*e.Contrast + 0.5
		g = (g-0.5)*e.Contrast + 0.5
		b = (b-0.5)*e.Contrast + 0.5

		avg := (r + g + b) / 3.0
		r = avg + (r-avg)*e.Saturation
		g = avg + (g-avg)*e.Saturation
		b = avg + (b-avg)*e.Saturation

		r = AdjustBrightness(r, e.Brightness)
		g = AdjustBrightness(g, e.Brightness)
		b = AdjustBrightness(b, e.Brightness)

		r, g, b = RotateHue(r, g, b, e.Hue)

		r = ApplyGamma(r, e.Gamma)
		g = ApplyGamma(g, e.Gamma)
		b = ApplyGamma(b, e.Gamma)

		r = AdjustShadowStrength(r, e.ShadowStrength)
		g = AdjustShadowStrength(g, e.ShadowStrength)
		b = AdjustShadowStrength(b, e.ShadowStrength)

		r = Clamp(r, 0, 1)
		g = Clamp(g, 0, 1)
		b = Clamp(b, 0, 1)

		return color.NRGBA{
			R: uint8(r * 255),
			G: uint8(g * 255),
			B: uint8(b * 255),
			A: c.A,
		}
	})
}
