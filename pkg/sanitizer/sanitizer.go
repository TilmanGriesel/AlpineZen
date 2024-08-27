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
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	// Verify file size; avoid processing very large files
	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to get file info: %w", err)
	}

	const maxSize = 10 * 1024 * 1024 // 10 MB limit
	if fileInfo.Size() > maxSize {
		file.Close()
		return fmt.Errorf("file is too large: %d bytes", fileInfo.Size())
	}

	// Ensure it's a valid image format
	img, format, err := image.Decode(file)
	file.Close() // Close the file as soon as it's no longer needed
	if err != nil {
		return fmt.Errorf("invalid image format or corrupt image: %w", err)
	}

	// Validate decoded image to ensure it's not nil
	if img == nil {
		return fmt.Errorf("decoded image is nil")
	}

	// Delete original file
	if err := os.Remove(filepath.Clean(filePath)); err != nil {
		return fmt.Errorf("failed to delete the original file: %w", err)
	}

	// Create sanitized file
	outFile, err := os.Create(filepath.Clean(filePath))
	if err != nil {
		return fmt.Errorf("failed to create new file: %w", err)
	}
	defer outFile.Close()

	// Ensure correct encoding
	switch strings.ToLower(format) {
	case "jpeg":
		err = jpeg.Encode(outFile, img, nil)
	case "png":
		err = png.Encode(outFile, img)
	default:
		return fmt.Errorf("unsupported image format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to re-encode image: %w", err)
	}

	if err := outFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync new file to disk: %w", err)
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
