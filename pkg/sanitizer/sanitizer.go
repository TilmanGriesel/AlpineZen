// SPDX-FileCopyrightText: 2025 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later AND LicenseRef-AlpineZen-Trademark

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

	cleanedPath := filepath.Clean(filePath)
	file, err := os.Open(cleanedPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	// Ensure the input file is closed properly
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close input file: %w", cerr)
		}
	}()

	// Verify file size; avoid processing very large files
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	const maxSize = 10 * 1024 * 1024 // 10 MB limit
	if fileInfo.Size() > maxSize {
		return fmt.Errorf("file is too large: %d bytes", fileInfo.Size())
	}

	// Ensure it's a valid image format
	img, format, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("invalid image format or corrupt image: %w", err)
	}

	// Validate decoded image to ensure it's not nil
	if img == nil {
		return fmt.Errorf("decoded image is nil")
	}

	// Close the input file before deleting
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close input file: %w", err)
	}

	// Delete original file
	if err := os.Remove(cleanedPath); err != nil {
		return fmt.Errorf("failed to delete the original file: %w", err)
	}

	// Create sanitized file
	outFile, err := os.Create(cleanedPath)
	if err != nil {
		return fmt.Errorf("failed to create new file: %w", err)
	}

	// Ensure the output file is closed properly
	defer func() {
		if cerr := outFile.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close output file: %w", cerr)
		}
	}()

	// Ensure correct encoding
	switch strings.ToLower(format) {
	case "jpeg":
		if encodeErr := jpeg.Encode(outFile, img, nil); encodeErr != nil {
			return fmt.Errorf("failed to encode JPEG image: %w", encodeErr)
		}
	case "png":
		if encodeErr := png.Encode(outFile, img); encodeErr != nil {
			return fmt.Errorf("failed to encode PNG image: %w", encodeErr)
		}
	default:
		return fmt.Errorf("unsupported image format: %s", format)
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
