// SPDX-FileCopyrightText: 2025 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later AND LicenseRef-AlpineZen-Trademark

package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/TilmanGriesel/AlpineZen/pkg/util"
)

type Janitor struct {
	FileType string
}

func NewJanitor(fileType string) *Janitor {
	return &Janitor{
		FileType: fileType,
	}
}

// DeepClean method deletes specified directory after validating path
func (j *Janitor) DeepClean(directory string) error {
	validPath, err := j.validatePath(directory)
	if !validPath {
		return err
	}

	info, err := os.Stat(directory)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", directory)
	}
	if err != nil {
		return fmt.Errorf("error checking directory: %v", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", directory)
	}

	absDir, err := filepath.Abs(directory)
	if err != nil {
		return fmt.Errorf("error getting absolute path: %v", err)
	}

	// Prevent deletion of important system directories
	if absDir == "/" || absDir == "/home" || absDir == "/usr" || absDir == "/etc" || absDir == "/var" {
		return fmt.Errorf("refusing to remove important system directory: %s", absDir)
	}

	err = os.RemoveAll(directory)
	if err != nil {
		return fmt.Errorf("error removing directory: %v", err)
	}

	return nil
}

// WipeThrough method removes files in directory exceeding maxFiles count, keeping most recent ones
func (j *Janitor) WipeThrough(directory string, maxFiles int) error {
	validPath, err := j.validatePath(directory)
	if !validPath {
		return fmt.Errorf("invalid path: %w", err)
	}

	files, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	var dirFiles []os.DirEntry
	for _, file := range files {
		if strings.HasSuffix(file.Name(), j.FileType) {
			dirFiles = append(dirFiles, file)
		}
	}

	if len(dirFiles) <= maxFiles {
		return nil
	}

	sort.Slice(dirFiles, func(i, j int) bool {
		fileInfoI, _ := dirFiles[i].Info()
		fileInfoJ, _ := dirFiles[j].Info()
		return fileInfoI.ModTime().After(fileInfoJ.ModTime())
	})

	for i := maxFiles; i < len(dirFiles); i++ {
		fileToRemove := filepath.Join(directory, dirFiles[i].Name())
		if err := os.Remove(fileToRemove); err != nil {
			return fmt.Errorf("failed to remove file %s: %w", fileToRemove, err)
		}
	}

	return nil
}

func (j *Janitor) validatePath(path string) (bool, error) {
	appDirName := util.GetAppDirName()
	if !strings.Contains(path, appDirName) {
		return false, fmt.Errorf("directory path must include '%s': %s", appDirName, path)
	}
	return true, nil
}
