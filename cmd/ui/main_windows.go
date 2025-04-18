//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/TilmanGriesel/AlpineZen/pkg/util"
	"github.com/tailscale/walk"
)

var (
	notifyIcon    *walk.NotifyIcon
	blurMenuItem  *walk.Action
	clockMenuItem *walk.Action
	menuItems     []*walk.Action
)

func main() {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	pathCli = filepath.Clean(filepath.Join(exeDir, "AlpineZenHelper.exe"))

	if _, err = os.Stat(pathCli); os.IsNotExist(err) {
		log.Fatalf("alpinezen-cli not found at expected path: %s", pathCli)
	}

	launchProcess(pathCli, false, "--prepare")

	app, err := walk.InitApp()
	if err != nil {
		log.Fatal(err)
	}

	settings, err := loadSettings()
	if err != nil {
		fmt.Println("Error loading settings:", err)
	}
	appSettings = settings

	setupSignalHandler()

	_, err = walk.NewMainWindow()
	if err != nil {
		log.Fatal(err)
	}

	notifyIcon, err = walk.NewNotifyIcon()
	if err != nil {
		log.Fatal(err)
	}
	defer notifyIcon.Dispose()

	iconPath := filepath.Join(exeDir, "MenuBarIcon.ico")
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		iconPath = filepath.Join(exeDir, "default_icon.ico")
		if _, err := os.Stat(iconPath); os.IsNotExist(err) {
			log.Println("Warning: Default icon not found")
		}
	}
	icon, err := walk.Resources.Icon(iconPath)
	if err != nil {
		icon, err = walk.NewIconFromSysDLL("shell32.dll", 17) // Folder icon as fallback
		if err != nil {
			log.Println("Warning: Could not load any icon:", err)
		}
	}
	if icon != nil {
		if err := notifyIcon.SetIcon(icon); err != nil {
			log.Println("Error setting icon:", err)
		}
	}
	
	if err := notifyIcon.SetToolTip("AlpineZen"); err != nil {
		log.Println("Error setting tooltip:", err)
	}
	
	if err := notifyIcon.SetVisible(true); err != nil {
		log.Println("Error setting visibility:", err)
	}

	initializeMenu()

	restartWithSettings(false)

	app.Run()
}

func setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("Received termination signal, shutting down...")
		terminateAllBackgroundProcesses()
		os.Exit(0)
	}()
}

func initializeMenu() {
	menu, _ := walk.NewMenu()
	defer menu.Dispose()

	createThemeMenuItems(menu)

	menu.Actions().Add(walk.NewSeparatorAction())

	// Blur option
	blurMenuItem = walk.NewAction()
	if err := blurMenuItem.SetText("Blur"); err != nil {
		log.Println("Error setting blur menu text:", err)
	}
	blurMenuItem.SetChecked(appSettings.SelectedType == "blur")
	blurMenuItem.Triggered().Attach(func() { toggleBlurState() })
	menu.Actions().Add(blurMenuItem)

	// Clock option
	clockMenuItem = walk.NewAction()
	if err := clockMenuItem.SetText("Show Clock"); err != nil {
		log.Println("Error setting clock menu text:", err)
	}
	clockMenuItem.SetChecked(appSettings.ShowClock)
	clockMenuItem.Triggered().Attach(func() { toggleShowClockState() })
	menu.Actions().Add(clockMenuItem)

	menu.Actions().Add(walk.NewSeparatorAction())

	// Log and quit options
	logAction := walk.NewAction()
	if err := logAction.SetText("Show Logfile"); err != nil {
		log.Println("Error setting logfile menu text:", err)
	}
	logAction.Triggered().Attach(openLogfiles)
	menu.Actions().Add(logAction)

	quitAction := walk.NewAction()
	if err := quitAction.SetText("Quit"); err != nil {
		log.Println("Error setting quit menu text:", err)
	}
	quitAction.Triggered().Attach(func() {
		terminateAllBackgroundProcesses()
		walk.App().Exit(0)
	})
	menu.Actions().Add(quitAction)

	contextMenu := notifyIcon.ContextMenu()
	contextMenu.Actions().Clear()
	
	n := menu.Actions().Len()
	for i := 0; i < n; i++ {
		action := menu.Actions().At(i)
		contextMenu.Actions().Add(action)
	}
}

func createThemeMenuItems(menu *walk.Menu) {
	items, err := util.GetBasecampRepoItemNames()
	if err != nil {
		fmt.Println("Error getting default repo items:", err)
		return
	}

	menuItems = make([]*walk.Action, 0, len(items))

	for _, itemName := range items {
		name := itemName

		action := walk.NewAction()
		if err := action.SetText(name); err != nil {
			log.Println("Error setting menu item text:", err)
			continue
		}
		action.SetChecked(strings.ToLower(name) == appSettings.SelectedName)
		
		action.Triggered().Attach(func() {
			handleMenuItemClick(name)
		})
		
		menu.Actions().Add(action)
		menuItems = append(menuItems, action)
	}
}

func handleMenuItemClick(itemName string) {
	// Uncheck all items
	for _, item := range menuItems {
		item.SetChecked(false)
	}
	
	// Find and check the selected item
	for _, item := range menuItems {
		text := item.Text()
		if text == itemName {
			item.SetChecked(true)
			break
		}
	}
	
	appSettings.SelectedName = strings.ToLower(itemName)
	saveSettingsAndRestart()
}

func toggleBlurState() {
	if blurMenuItem.Checked() {
		blurMenuItem.SetChecked(false)
		appSettings.SelectedType = "default"
	} else {
		blurMenuItem.SetChecked(true)
		appSettings.SelectedType = "blur"
	}
	saveSettingsAndRestart()
}

func toggleShowClockState() {
	if clockMenuItem.Checked() {
		clockMenuItem.SetChecked(false)
		appSettings.ShowClock = false
	} else {
		clockMenuItem.SetChecked(true)
		appSettings.ShowClock = true
	}
	saveSettingsAndRestart()
}

func openLogfiles() {
	path, err := util.GetAppDirPath()
	if err != nil {
		fmt.Println("Error getting app directory path:", err)
		return
	}

	logFilePath := filepath.Join(path, "log", "alpinezen_cli.log")
	
	cmd := exec.Command("cmd", "/c", "start", logFilePath)
	err = cmd.Run()
	
	if err != nil {
		fmt.Println("Error opening log files:", err, logFilePath)
		notifyIcon.ShowInfo("AlpineZen", "Error opening log file: "+err.Error())
	} else {
		fmt.Println("Log files opened successfully.")
	}
}