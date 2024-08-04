//go:build darwin
// +build darwin

package main

import (
	"C"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/TilmanGriesel/AlpineZen/pkg/util"
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"
	"github.com/progrium/darwinkit/objc"
)

func main() {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	pathCli = filepath.Clean(filepath.Join(exeDir, "AlpineZenHelper"))

	// Verify CLI exists
	if _, err = os.Stat(pathCli); os.IsNotExist(err) {
		log.Fatalf("alpinezen-cli not found at expected path: %s", pathCli)
	}

	// Prepare config structure and default repository
	launchProcess(pathCli, false, "--prepare")

	runtime.LockOSThread()

	// Load and store settings
	appSettings, err = loadSettings()
	if err != nil {
		fmt.Println("Error loading settings:", err)
	}

	app := appkit.Application_SharedApplication()
	delegate := &appkit.ApplicationDelegate{}

	// Setup system notifications
	dnc := foundation.DistributedNotificationCenter_NotificationCenterForType(foundation.LocalNotificationCenterType)
	dnc.AddObserverForNameObjectQueueUsingBlock("com.apple.screenIsLocked", nil, foundation.OperationQueue_MainQueue(), func(notification foundation.Notification) {
		log.Println("screen is locked")
		restartWithSettings(true)
	})
	dnc.AddObserverForNameObjectQueueUsingBlock("com.apple.screenIsUnlocked", nil, foundation.OperationQueue_MainQueue(), func(notification foundation.Notification) {
		log.Println("screen is unlocked")
		restartWithSettings(false)
	})

	delegate.SetApplicationDidFinishLaunching(func(foundation.Notification) {
		setSystemBar(app)
		app.SetActivationPolicy(appkit.ApplicationActivationPolicyRegular)
		app.ActivateIgnoringOtherApps(true)

		restartWithSettings(false)
	})

	delegate.SetApplicationShouldTerminate(func(appkit.Application) appkit.ApplicationTerminateReply {
		terminateAllBackgroundProcesses()
		return appkit.TerminateNow
	})

	setupSignalHandler()

	app.SetDelegate(delegate)
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

func setSystemBar(app appkit.Application) {
	item := appkit.StatusBar_SystemStatusBar().StatusItemWithLength(appkit.VariableStatusItemLength)
	objc.Retain(&item)

	img := appkit.Image_ImageNamed("MenuBarIcon")
	if img.IsNil() || !img.IsValid() {
		img = appkit.Image_ImageWithSystemSymbolNameAccessibilityDescription("mountain.2.circle.fill", "")
		if img.IsNil() || !img.IsValid() {
			fmt.Println("Error: Unable to find both custom image and system symbol.")
			return
		}
	}

	// DEBUG
	// img = appkit.Image_ImageWithSystemSymbolNameAccessibilityDescription("mountain.2.circle.fill", "")

	item.Button().SetImage(img)

	menu := appkit.NewMenuWithTitle("main")

	initializeMenuItems(menu)

	item.SetMenu(menu)
}

func initializeMenuItems(menu appkit.Menu) {
	var blurItem appkit.MenuItem
	blurItem = appkit.NewMenuItemWithAction("Blur", "", func(sender objc.Object) {
		toggleBlurState(&blurItem)
	})

	if appSettings.SelectedType == "blur" {
		blurItem.SetState(appkit.ControlStateValueOn)
	} else {
		blurItem.SetState(appkit.ControlStateValueOff)
	}

	var clockItem appkit.MenuItem
	clockItem = appkit.NewMenuItemWithAction("Show Clock", "", func(sender objc.Object) {
		toggleShowClockState(&clockItem)
	})

	if appSettings.ShowClock {
		clockItem.SetState(appkit.ControlStateValueOn)
	} else {
		clockItem.SetState(appkit.ControlStateValueOff)
	}

	createMenuItems(menu)

	menu.AddItem(appkit.MenuItem_SeparatorItem())
	menu.AddItem(blurItem)
	menu.AddItem(clockItem)

	menu.AddItem(appkit.MenuItem_SeparatorItem())
	menu.AddItem(appkit.NewMenuItemWithAction("Show Logfile", "h", func(sender objc.Object) { openLogfiles() }))
	menu.AddItem(appkit.NewMenuItemWithAction("Quit", "q", func(sender objc.Object) { quitApp(appkit.Application_SharedApplication()) }))
}

func toggleBlurState(menuItem *appkit.MenuItem) {
	if menuItem.State() == appkit.ControlStateValueOn {
		menuItem.SetState(appkit.ControlStateValueOff)
		appSettings.SelectedType = "default"
	} else {
		menuItem.SetState(appkit.ControlStateValueOn)
		appSettings.SelectedType = "blur"
	}
	saveSettingsAndRestart()
}

func toggleShowClockState(menuItem *appkit.MenuItem) {
	if menuItem.State() == appkit.ControlStateValueOn {
		menuItem.SetState(appkit.ControlStateValueOff)
		appSettings.ShowClock = false
	} else {
		menuItem.SetState(appkit.ControlStateValueOn)
		appSettings.ShowClock = true
	}
	saveSettingsAndRestart()
}

func createMenuItems(menu appkit.Menu) []appkit.MenuItem {
	items, err := util.GetBasecampRepoItemNames()
	if err != nil {
		fmt.Println("Error getting default repo items:", err)
		return nil
	}

	var menuItems []appkit.MenuItem

	for _, itemName := range items {
		var item appkit.MenuItem
		item = appkit.NewMenuItemWithAction(itemName, "", func(sender objc.Object) {
			handleMenuItemClick(&menuItems, &item, itemName)
		})

		if appSettings.SelectedName == strings.ToLower(itemName) {
			item.SetState(appkit.ControlStateValueOn)
		} else {
			item.SetState(appkit.ControlStateValueOff)
		}

		menu.AddItem(item)
		menuItems = append(menuItems, item)
	}

	return menuItems
}

func handleMenuItemClick(menuItems *[]appkit.MenuItem, item *appkit.MenuItem, itemName string) {
	for _, mi := range *menuItems {
		mi.SetState(appkit.ControlStateValueOff)
	}
	item.SetState(appkit.ControlStateValueOn)
	appSettings.SelectedName = strings.ToLower(itemName)
	saveSettingsAndRestart()
}

func openLogfiles() {
	path, err := util.GetAppDirPath()

	if err != nil {
		fmt.Println("Error getting app directory path:", err)
		return
	}

	logFilePath := filepath.Join(path, "log", "alpinezen_cli.log")
	err = exec.Command("open", logFilePath).Run()
	if err != nil {
		fmt.Println("Error opening log files:", err)
	} else {
		fmt.Println("Log files opened successfully.")
	}
}

func quitApp(app appkit.Application) {
	terminateAllBackgroundProcesses()
	app.Terminate(nil)
}
