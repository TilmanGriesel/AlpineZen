// SPDX-FileCopyrightText: 2024 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later

package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"math/big"
	"strings"

	"net/http"
	"os"
	"path/filepath"
)

const (
	HashLength = 16
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.181 Safari/537.36",                                     // Chrome 66 on Windows 10 (2018)
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_4) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/11.1 Safari/605.1.15",                                   // Safari 11.1 on macOS 10.13 (2018)
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.75 Safari/537.36",                                                // Chrome 68 on Linux (2018)
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:61.0) Gecko/20100101 Firefox/61.0",                                                                          // Firefox 61 on Windows 10 (2018)
	"Mozilla/5.0 (iPhone; CPU iPhone OS 11_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/11.0 Mobile/15E148 Safari/604.1",                 // Safari on iPhone iOS 11.4 (2018)
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36",                                       // Chrome 67 on Windows 7 (2018)
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/11.1.2 Safari/605.1.15",                                 // Safari 11.1.2 on macOS 10.12 (2018)
	"Mozilla/5.0 (Linux; Android 8.0.0; Pixel 2 XL Build/OPM4.171019.021.P1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.87 Mobile Safari/537.36", // Chrome 67 on Android 8.0 (2018)
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36",                                     // Chrome 65 on Windows 10 (2018)
	"Mozilla/5.0 (iPad; CPU OS 11_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/11.0 Mobile/15E148 Safari/604.1",                          // Safari on iPad iOS 11.3 (2018)
	"Mozilla/5.0 (X11; Linux x86_64) Gecko/20100101 Firefox/60.0",                                                                                             // Firefox 60 on Linux (2018)
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.0 Safari/605.1.15",                                     // Safari 12 on macOS 10.14 (2018)
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36",                                      // Chrome 70 on Windows 10 (2018)
	"Mozilla/5.0 (Linux; Android 9; Pixel 3 XL Build/PQ1A.181105.017.A1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.110 Mobile Safari/537.36",    // Chrome 70 on Android 9 (2018)
	"Mozilla/5.0 (iPhone; CPU iPhone OS 12_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/12.0 Mobile/15E148 Safari/604.1",                 // Safari on iPhone iOS 12.1 (2018)
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36",                                      // Chrome 69 on Windows 7 (2018)
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/11.1.2 Safari/605.1.15",                                 // Safari 11.1.2 on macOS 10.13 (2018)
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:62.0) Gecko/20100101 Firefox/62.0",                                                                            // Firefox 62 on Ubuntu Linux (2018)
	"Mozilla/5.0 (Linux; Android 8.1.0; Pixel 2 Build/OPM2.171019.029) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/68.0.3440.91 Mobile Safari/537.36",       // Chrome 68 on Android 8.1 (2018)
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:63.0) Gecko/20100101 Firefox/63.0",                                                                          // Firefox 63 on Windows 10 (2018)
}

func RandomUserAgent() string {
	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(userAgents))))
	if err != nil {
		log.Fatalf("failed to generate random number: %v", err)
	}
	return userAgents[index.Int64()]
}

func DownloadImage(url, path string, honest bool) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	userAgent := "AlpineZen-Wallpaper/1.0"
	if !honest {
		userAgent = RandomUserAgent()
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	file, err := os.Create(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("failed to create file for downloaded image: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to copy image data to file: %w", err)
	}

	return nil
}

func HashSHA256(data string) string {
	hash := sha256.New()
	hash.Write([]byte(data))
	return hex.EncodeToString(hash.Sum(nil))
}

func GetAppDirName() string {
	return ".alpinezen_wallpaper"
}

func GetAppDirPath() (string, error) {
	usr, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Clean(filepath.Join(usr, GetAppDirName())), nil
}

func GetDefaultRepoPath() (string, error) {
	path, err := GetAppDirPath()
	if err != nil {
		return "", err
	}

	return filepath.Clean(filepath.Join(path, "repos")), nil
}

func GetBasecampRepoItemNames() ([]string, error) {
	path, err := GetDefaultRepoPath()
	if err != nil {
		return nil, err
	}

	path = filepath.Join(path, "AlpineZen-Basecamp-main")

	entries, err := os.ReadDir(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	var repoItemNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			folderName := strings.Title(entry.Name())
			repoItemNames = append(repoItemNames, folderName)
		}
	}

	return repoItemNames, nil
}

func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(filepath.Clean(src))
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	err = destFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func LoadImageFile(path string) (image.Image, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func GenerateShortHash(input, seed string) string {
	fullHash := HashSHA256(input + seed)
	return fullHash[:HashLength]
}

func ParseHexColor(s string) (color.RGBA, error) {
	c := color.RGBA{A: 0xff}

	if s[0] == '#' {
		s = s[1:]
	}

	switch len(s) {
	case 6: // #RRGGBB
		bytes, err := hex.DecodeString(s)
		if err != nil {
			return c, err
		}
		c.R = bytes[0]
		c.G = bytes[1]
		c.B = bytes[2]
	default:
		return c, fmt.Errorf("invalid length, must be 6 characters")
	}

	return c, nil
}
