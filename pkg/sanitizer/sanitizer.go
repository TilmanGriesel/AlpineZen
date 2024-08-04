// SPDX-FileCopyrightText: 2024 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later

package sanitizer

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

func SanitizeImage(filePath string) error {
	// Check if filePath is empty
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Open file securely
	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Verify file size to avoid processing very large files
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	const maxSize = 10 * 1024 * 1024 // 10 MB limit
	if fileInfo.Size() > maxSize {
		return fmt.Errorf("file is too large: %d bytes", fileInfo.Size())
	}

	// Decode image to ensure it's a valid image format
	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("invalid image format or corrupt image: %w", err)
	}

	// Validate decoded image to ensure it's not nil
	if img == nil {
		return fmt.Errorf("decoded image is nil")
	}

	// Create a temporary file path securely
	tempSanitizedFilePath := filePath + ".sanitized"
	outFile, err := os.Create(filepath.Clean(tempSanitizedFilePath))
	if err != nil {
		return fmt.Errorf("failed to create sanitized file: %w", err)
	}
	defer outFile.Close()

	// Ensure correct format is used for encoding
	switch strings.ToLower(format) {
	case "jpeg":
		err = jpeg.Encode(outFile, img, nil) // use default options for JPEG
	case "png":
		err = png.Encode(outFile, img)
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to re-encode image: %w", err)
	}

	// Ensure data is written to disk before replacing original file
	if err := outFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync sanitized file to disk: %w", err)
	}

	// Replace original file with sanitized version
	err = os.Rename(tempSanitizedFilePath, filePath)
	if err != nil {
		return fmt.Errorf("failed to replace original image with sanitized version: %w", err)
	}

	return nil
}

func SanitizeArchivePath(dir, target string) (string, error) {
	// Ref: https://security.snyk.io/research/zip-slip-vulnerability

	validPath := filepath.Join(dir, target)
	if strings.HasPrefix(validPath, filepath.Clean(dir)+string(os.PathSeparator)) {
		return validPath, nil
	}
	return "", fmt.Errorf("content filepath is tainted: %s", target)
}
