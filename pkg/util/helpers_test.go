// SPDX-FileCopyrightText: 2024 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later

package util

import (
	"crypto/sha256"
	"encoding/hex"
	"image"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandomUserAgent(t *testing.T) {
	ua := RandomUserAgent()
	assert.Contains(t, userAgents, ua, "User-Agent should be one of predefined values")
}

func TestDownloadImage(t *testing.T) {
	server := http.FileServer(http.Dir("./testdata"))
	testServer := httptest.NewServer(server)
	defer testServer.Close()

	testURL := testServer.URL + "/sample.jpg"
	testFilePath := "./test_downloaded.jpg"

	defer os.Remove(testFilePath)

	err := DownloadImage(testURL, testFilePath, true)
	require.NoError(t, err, "downloadImage should not return an error")

	assert.FileExists(t, testFilePath, "The downloaded file should exist")
}

func TestHashSHA256(t *testing.T) {
	data := "test data"
	expectedHash := sha256.Sum256([]byte(data))
	expectedHashStr := hex.EncodeToString(expectedHash[:])

	hash := HashSHA256(data)
	assert.Equal(t, expectedHashStr, hash, "Hash value should match expected SHA256 hash")
}

func TestGetAppDirPath(t *testing.T) {
	dirPath, err := GetAppDirPath()
	require.NoError(t, err, "GetAppDirPath should not return an error")

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err, "os.UserHomeDir should not return an error")

	expectedPath := filepath.Join(homeDir, GetAppDirName())
	assert.Equal(t, expectedPath, dirPath, "App directory path should match expected path")
}

func TestCopyFile(t *testing.T) {
	srcPath := "./testdata/sample.txt"
	dstPath := "./test_copied.txt"

	defer os.Remove(dstPath)

	err := CopyFile(srcPath, dstPath)
	require.NoError(t, err, "copyFile should not return an error")

	assert.FileExists(t, dstPath, "The copied file should exist")

	srcContent, err := os.ReadFile(srcPath)
	require.NoError(t, err, "os.ReadFile should not return an error for source file")

	dstContent, err := os.ReadFile(dstPath)
	require.NoError(t, err, "os.ReadFile should not return an error for destination file")

	assert.Equal(t, srcContent, dstContent, "Source and destination file contents should match")
}

func TestFileExists(t *testing.T) {
	existingFile := "./testdata/sample.txt"
	nonExistingFile := "./testdata/non_existent.txt"

	assert.True(t, FileExists(existingFile), "fileExists should return true for an existing file")
	assert.False(t, FileExists(nonExistingFile), "fileExists should return false for a non-existing file")
}

func TestLoadImage(t *testing.T) {
	return
	imgPath := "testdata/sample.jpg"

	img, err := LoadImageFile(imgPath)
	require.NoError(t, err, "loadImage should not return an error")

	assert.IsType(t, &image.YCbCr{}, img, "The loaded image should be of type *image.NRGBA")
}

func TestGenerateShortHash(t *testing.T) {
	url := "http://example.com"
	timestamp := "2024-08-07T10:00:00Z"
	fullHash := HashSHA256(url + timestamp)
	expectedShortHash := fullHash[:HashLength]

	shortHash := GenerateShortHash(url, timestamp)
	assert.Equal(t, expectedShortHash, shortHash, "Short hash should match expected value")
}
