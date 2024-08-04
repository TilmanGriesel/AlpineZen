// SPDX-FileCopyrightText: 2025 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later AND LicenseRef-AlpineZen-Trademark

package repository

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/TilmanGriesel/AlpineZen/pkg/sanitizer"
	"github.com/sirupsen/logrus"
)

const (
	maxCompressedFileSize = 10 * 1024 * 1024  // 10 MB
	maxUncompressedSize   = 100 * 1024 * 1024 // 100 MB
	maxSingleFileSize     = 50 * 1024 * 1024  // 50 MB
)

func GetRepoFolderName(repoURL string) string {
	re := regexp.MustCompile(`([^/]+)/archive/refs/heads/([^/]+)\.zip$`)
	matches := re.FindStringSubmatch(repoURL)
	if len(matches) != 3 {
		return ""
	}
	return fmt.Sprintf("%s-%s", matches[1], matches[2])
}

func DownloadAndExtractZip(url, path string) error {
	if err := os.MkdirAll(filepath.Clean(path), 0750); err != nil {
		return fmt.Errorf("failed to create directory: %s %w", path, err)
	}

	config := DefaultRetryConfig()
	config.Logger = logrus.StandardLogger()

	downloadOperation := func() error {
		client := &http.Client{
			Timeout: 30 * time.Second,
		}
		resp, err := client.Get(url)
		if err != nil {
			return fmt.Errorf("failed to download zip file: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to download zip file: status code %d", resp.StatusCode)
		}

		buf := new(bytes.Buffer)
		_, err = io.CopyN(buf, resp.Body, maxCompressedFileSize+1)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read zip file: %w", err)
		}

		if buf.Len() > maxCompressedFileSize {
			return fmt.Errorf("downloaded file is too large")
		}

		if err = extractZip(buf, path); err != nil {
			return fmt.Errorf("failed to extract zip file: %w", err)
		}

		return nil
	}

	return RetryWithExponentialBackoff(downloadOperation, config)
}

func extractZip(buf *bytes.Buffer, extractTo string) error {
	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %w", err)
	}

	absExtractTo, err := filepath.Abs(extractTo)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path for extraction directory: %w", err)
	}

	var totalSize uint64
	for _, f := range zipReader.File {
		totalSize += f.UncompressedSize64
		if totalSize > maxUncompressedSize {
			return fmt.Errorf("total uncompressed size exceeds limit")
		}

		if f.UncompressedSize64 > maxSingleFileSize {
			return fmt.Errorf("file %s exceeds max single file size limit", f.Name)
		}

		sanitizedPath, err := sanitizer.SanitizeArchivePath(absExtractTo, f.Name)
		if err != nil {
			return fmt.Errorf("sanitization error: %w", err)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(sanitizedPath, f.Mode()); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", sanitizedPath, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(sanitizedPath), 0750); err != nil {
			return fmt.Errorf("failed to create directory for file %s: %w", sanitizedPath, err)
		}

		if err = extractFile(f, sanitizedPath); err != nil {
			return err
		}
	}

	return nil
}

func extractFile(f *zip.File, path string) error {
	const maxFileSize = 10 * 1024 * 1024

	outFile, err := os.OpenFile(filepath.Clean(path), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return fmt.Errorf("failed to open file %s for writing: %w", path, err)
	}
	defer outFile.Close()

	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("failed to open zip entry %s: %w", f.Name, err)
	}
	defer rc.Close()

	if f.UncompressedSize64 > uint64(maxFileSize) {
		return fmt.Errorf("uncompressed size too large: %d", f.UncompressedSize64)
	}

	limitedReader := &io.LimitedReader{
		R: rc,
		N: int64(maxFileSize),
	}

	if _, err = io.Copy(outFile, limitedReader); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	if limitedReader.N <= 0 {
		return fmt.Errorf("decompressed file %s exceeds the allowed size limit", path)
	}

	return nil
}
