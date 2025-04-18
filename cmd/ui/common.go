package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/TilmanGriesel/AlpineZen/pkg/util"
)

var (
	backgroundProcesses []*exec.Cmd
	mu                  sync.Mutex
	appSettings         AppSettings
	pathCli             string
)

type AppSettings struct {
	SelectedName string `yaml:"selectedName"`
	SelectedType string `yaml:"selectedType"`
	ShowClock    bool   `yaml:"showClock"`
}

func DefaultSettings() AppSettings {
	return AppSettings{
		SelectedName: "fellhorn",
		SelectedType: "default",
		ShowClock:    true,
	}
}

func getSettingsFilePath() string {
	appDir, _ := util.GetAppDirPath()
	return filepath.Join(appDir, "config", "gui.yaml")
}

func loadSettings() (AppSettings, error) {
	var settings AppSettings
	path := getSettingsFilePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = saveSettings(DefaultSettings())
		if err != nil {
			return settings, err
		}
	}

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return settings, err
	}

	err = json.Unmarshal(data, &settings)
	return settings, err
}

func saveSettingsAndRestart() error {
	err := saveSettings(appSettings)
	if err != nil {
		return err
	}
	restartWithSettings(false)
	return nil
}

func saveSettings(settings AppSettings) error {
	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	filePath := getSettingsFilePath()
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0750)
		if err != nil {
			return err
		}
	}

	return os.WriteFile(filePath, data, 0600)
}

func terminateBackgroundProcess(pid int) {
	mu.Lock()
	defer mu.Unlock()

	for i, cmd := range backgroundProcesses {
		if cmd.Process != nil && cmd.Process.Pid == pid {
			fmt.Println("Terminating background process with PID:", pid)
			if err := cmd.Process.Kill(); err != nil {
				fmt.Println("Error killing background process:", err)
			} else {
				fmt.Println("Background process terminated.")
				backgroundProcesses = append(backgroundProcesses[:i], backgroundProcesses[i+1:]...)
			}
			return
		}
	}
	fmt.Println("No background process found with PID:", pid)
}
