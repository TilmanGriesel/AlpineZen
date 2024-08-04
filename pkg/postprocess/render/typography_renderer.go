// SPDX-FileCopyrightText: 2025 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later AND LicenseRef-AlpineZen-Trademark

package render

import (
	_ "embed"
	"image"
	"image/color"
	"image/draw"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/TilmanGriesel/AlpineZen/pkg/postprocess"
	"github.com/disintegration/imaging"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// Embed default font files
//
//go:embed assets/fonts/Lora/Lora-Regular.ttf
var defaultFontRegular []byte

//go:embed assets/fonts/Lora/Lora-Italic.ttf
var defaultFontItalic []byte

//go:embed assets/fonts/Lora/Lora-Bold.ttf
var defaultFontBold []byte

type HorizontalAlignment int
type VerticalAlignment int
type FontStyle int

const (
	AlignCenter HorizontalAlignment = iota
	AlignLeft
	AlignRight
)

const (
	AlignMiddle VerticalAlignment = iota
	AlignTop
	AlignBottom
)

const (
	Regular FontStyle = iota
	Italic
	Bold
)

type FontConfig struct {
	FontPath   string
	Size       float64
	DPI        float64
	Color      color.Color
	Style      FontStyle
	MinOpacity float64
	MaxOpacity float64
	Position   FontPositionConfig
	TimeFormat string // Added to make time format configurable
}

type FontPositionConfig struct {
	HorizontalAlignment    HorizontalAlignment
	VerticalAlignment      VerticalAlignment
	PaddingTop             int
	PaddingBottom          int
	PaddingLeft            int
	PaddingRight           int
	VerticalCenterOffset   int
	HorizontalCenterOffset int
}

func getFontBytes(customPath string, style ...FontStyle) []byte {
	if customPath != "" {
		fontBytes, err := os.ReadFile(filepath.Clean(customPath))
		if err != nil {
			log.Fatalf("failed to read font file: %v", err)
		}
		return fontBytes
	}

	var selectedStyle FontStyle
	if len(style) > 0 {
		selectedStyle = style[0]
	} else {
		selectedStyle = Regular
	}

	switch selectedStyle {
	case Bold:
		return defaultFontBold
	case Italic:
		return defaultFontItalic
	default:
		return defaultFontRegular
	}
}

func DrawTimeOnImage(img image.Image, fontConfig FontConfig) image.Image {
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, image.Point{}, draw.Src)

	// Get font bytes from embedded font data
	fontBytes := getFontBytes(fontConfig.FontPath, fontConfig.Style)

	// Parse font
	fnt, err := truetype.Parse(fontBytes)
	if err != nil {
		log.Fatalf("failed to parse font: %v", err)
	}

	face := truetype.NewFace(fnt, &truetype.Options{
		Size: fontConfig.Size,
		DPI:  fontConfig.DPI,
	})

	// Use configurable time format
	currentTime := time.Now().Format(fontConfig.TimeFormat)

	d := &font.Drawer{
		Dst:  rgba,
		Src:  image.NewUniform(fontConfig.Color),
		Face: face,
	}

	textWidth := d.MeasureString(currentTime).Ceil()
	textHeight := face.Metrics().Height.Ceil()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	var centerX, centerY int

	// Horizontal Alignment with Offset
	switch fontConfig.Position.HorizontalAlignment {
	case AlignLeft:
		centerX = fontConfig.Position.PaddingLeft
	case AlignRight:
		centerX = imgWidth - textWidth - fontConfig.Position.PaddingRight
	default: // center
		centerX = (imgWidth-textWidth)/2 + fontConfig.Position.HorizontalCenterOffset
	}

	// Vertical Alignment with Offset
	switch fontConfig.Position.VerticalAlignment {
	case AlignTop:
		centerY = fontConfig.Position.PaddingTop + textHeight
	case AlignBottom:
		centerY = imgHeight - fontConfig.Position.PaddingBottom
	default: // center
		centerY = (imgHeight+textHeight)/2 + fontConfig.Position.VerticalCenterOffset
	}

	point := fixed.P(centerX, centerY)

	d.Dot = point
	d.DrawString(currentTime)

	textImage := image.NewRGBA(bounds)
	draw.Draw(textImage, bounds, rgba, image.Point{}, draw.Over)

	avgBrightness := postprocess.CalculateAverageBrightness(img)
	scaledOpacity := postprocess.CalculateScaledOpacity(avgBrightness, 0.0, 0.4, fontConfig.MinOpacity, fontConfig.MaxOpacity)
	blended := imaging.Overlay(img, textImage, image.Point{}, scaledOpacity)

	return blended
}
