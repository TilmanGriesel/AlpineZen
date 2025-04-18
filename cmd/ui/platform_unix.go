//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

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