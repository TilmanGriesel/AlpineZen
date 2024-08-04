// SPDX-FileCopyrightText: 2025 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later AND LicenseRef-AlpineZen-Trademark

package adjustment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClamp(t *testing.T) {
	tests := []struct {
		value    float64
		min      float64
		max      float64
		expected float64
	}{
		{value: 5.0, min: 0.0, max: 10.0, expected: 5.0},
		{value: -5.0, min: 0.0, max: 10.0, expected: 0.0},
		{value: 15.0, min: 0.0, max: 10.0, expected: 10.0},
	}

	for _, tt := range tests {
		clampedValue := Clamp(tt.value, tt.min, tt.max)
		assert.Equal(t, tt.expected, clampedValue, "Clamped value should match expected value")
	}
}
