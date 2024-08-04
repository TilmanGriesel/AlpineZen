// SPDX-FileCopyrightText: 2025 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later AND LicenseRef-AlpineZen-Trademark

//go:build !darwin
// +build !darwin

package wallpaper

import (
	"github.com/reujab/wallpaper"
)

func SetWallpaper(filepath string) error {
	return wallpaper.SetFromFile(filepath)
}
