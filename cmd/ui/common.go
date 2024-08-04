package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"

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

func terminateAllBackgroundProcesses() {
	mu.Lock()
	defer mu.Unlock()

	for _, cmd := range backgroundProcesses {
		if cmd.Process != nil {
			fmt.Println("Terminating background process with PID:", cmd.Process.Pid)
			if err := cmd.Process.Kill(); err != nil {
				fmt.Println("Error killing background process:", err)
			} else {
				fmt.Println("Background process terminated.")
			}
		}
	}
	backgroundProcesses = nil
}

func launchProcess(path string, background bool, args ...string) {
	cmd := exec.Command(path, args...)

	if background {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}
		fmt.Println("Launching process in background mode.")
	} else {
		fmt.Println("Launching process in foreground mode.")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting command:", err)
		return
	}

	if background {
		mu.Lock()
		backgroundProcesses = append(backgroundProcesses, cmd)
		mu.Unlock()
		fmt.Println("Background process started with PID:", cmd.Process.Pid)
	} else {
		fmt.Println("Foreground process started with PID:", cmd.Process.Pid)

		// Wait for the foreground process to finish
		if err := cmd.Wait(); err != nil {
			fmt.Println("Foreground process finished with error:", err)
		} else {
			fmt.Println("Foreground process finished successfully.")
		}
	}
}

func restartWithSettings(forceHideClock bool) {
	terminateAllBackgroundProcesses()

	args := []string{
		"--name", appSettings.SelectedName,
		"--type", appSettings.SelectedType,
	}

	if !appSettings.ShowClock || forceHideClock {
		args = append(args, "--clock-disable")
	}

	launchProcess(pathCli, true, args...)
}
