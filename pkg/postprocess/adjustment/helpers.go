// SPDX-FileCopyrightText: 2024 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later

package adjustment

import "math"

func AdjustBrightness(c float64, brightness float64) float64 {
	return Clamp(c+brightness, 0, 1)
}

func RotateHue(r, g, b, angle float64) (float64, float64, float64) {
	u := math.Cos(angle * math.Pi / 180.0)
	w := math.Sin(angle * math.Pi / 180.0)
	newR := (.299+.701*u+.168*w)*r + (.587-.587*u+.330*w)*g + (.114-.114*u-.497*w)*b
	newG := (.299-.299*u-.328*w)*r + (.587+.413*u+.035*w)*g + (.114-.114*u+.292*w)*b
	newB := (.299-.3*u+1.25*w)*r + (.587-.588*u-1.05*w)*g + (.114+.886*u-.203*w)*b
	return newR, newG, newB
}

func ApplyGamma(c, gamma float64) float64 {
	return math.Pow(c, 1.0/gamma)
}

func AdjustBlackPoint(c, blackPoint float64) float64 {
	return Clamp((c-blackPoint)/(1-blackPoint), 0, 1)
}

func AdjustWhitePoint(c, whitePoint float64) float64 {
	return Clamp(c*(1/whitePoint), 0, 1)
}

func AdjustShadowStrength(c, strength float64) float64 {
	return Clamp(c*strength, 0, 1)
}

func Clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
