// SPDX-FileCopyrightText: 2024 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later

package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/TilmanGriesel/AlpineZen/pkg/util"
	"github.com/sirupsen/logrus"
)

var (
	logger      *logrus.Logger
	initErr     error
	logFileName string
)

func init() {
	logger = logrus.New()
	logger.Out = os.Stdout
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}

func SetLogFileName(logFileName string) {
	path, err := util.GetAppDirPath()
	if err != nil || path == "" {
		initErr = fmt.Errorf("failed to get app directory path: %w", err)
		return
	}

	logFilePath := filepath.Clean(filepath.Join(path, "log", logFileName))

	if err := os.MkdirAll(filepath.Dir(logFilePath), 0750); err != nil {
		logger.WithError(err).Error("Failed to create log file directory")
		initErr = err
		return
	}

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		initErr = fmt.Errorf("failed to open log file: %w", err)
		return
	}

	logger.SetOutput(io.MultiWriter(os.Stdout, file))

	logger.WithField("logFileName", logFileName).Debug("Log file name set")
}

func GetLogger() *logrus.Logger {
	if initErr != nil {
		logger.WithError(initErr).Error("Logger initialization failed")
	}
	return logger
}
