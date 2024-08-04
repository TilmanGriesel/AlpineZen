// SPDX-FileCopyrightText: 2024 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetupDefaultRepository(t *testing.T) {
	test := struct {
		name            string
		args            []string
		expectedMessage string
		expectFatal     bool
	}{
		name: "Default configuration",
		args: []string{
			"--loglevel=3",
			"--runtime-headless=true",
			"--runtime-cpu-cores=" + strconv.Itoa(runtime.NumCPU())},
		expectedMessage: "Wallpaper update completed",
		expectFatal:     false,
	}

	t.Run(test.name, func(t *testing.T) {
		// Save original args and reset after test
		origArgs := os.Args
		defer func() { os.Args = origArgs }()

		// Set test args
		os.Args = append([]string{origArgs[0]}, test.args...)

		// Capture log output
		var logBuffer bytes.Buffer
		logger.SetOutput(io.MultiWriter(os.Stdout, &logBuffer))
		defer logger.SetOutput(os.Stderr)

		// Run application
		app := &Application{}

		go app.processFlags()

		// Allow some time for processing
		time.Sleep(15 * time.Second)

		// Check log output
		logOutput := logBuffer.String()
		if !strings.Contains(logOutput, test.expectedMessage) {
			t.Errorf("Expected log message %q not found in output: %s", test.expectedMessage, logOutput)
		}
	})
}

func TestProcessFlagsAndRun(t *testing.T) {
	configPath := filepath.Join("testdata", "config", "sample", "default.yaml")
	assert.FileExists(t, configPath, "Test config file should exist")

	app := &Application{}

	os.Args = []string{
		"alpinezen",
		"--config-path", configPath,
		"--runtime-cpu-cores", "2",
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Application panicked during flag processing: %v", r)
			}
		}()
		app.processFlags()
	}()

	time.Sleep(2 * time.Second)

	assert.NotNil(t, app.WallpaperManager, "WallpaperManager should be initialized after processFlags")
	assert.NotNil(t, app.UpdaterManager, "UpdaterManager should be initialized after processFlags")
}

func TestMainLoop(t *testing.T) {
	app := &Application{}

	// Simulate running main loop with a minimal setup
	go func() {
		app.processFlags()
	}()

	// Allow loop to run for a short period
	time.Sleep(2 * time.Second)

	// No specific assertions, but ensure no panics or errors occur
}
