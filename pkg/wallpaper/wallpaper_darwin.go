// SPDX-FileCopyrightText: 2025 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later AND LicenseRef-AlpineZen-Trademark

//go:build darwin
// +build darwin

package wallpaper

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"
	"github.com/progrium/darwinkit/objc"
)

var (
	once sync.Once
	app  appkit.Application
)

func initApp() {
	once.Do(func() {
		objc.WithAutoreleasePool(func() {
			app = appkit.Application_SharedApplication()
			app.SetActivationPolicy(appkit.ApplicationActivationPolicyRegular)
			app.ActivateIgnoringOtherApps(true)
		})
	})
}

func SetWallpaper(filepath string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var setError error

	initApp()

	objc.WithAutoreleasePool(func() {
		workspace := appkit.Workspace_SharedWorkspace()
		url := foundation.URL_FileURLWithPath(filepath)

		screens := appkit.Screen_Screens()
		for _, screen := range screens {
			var errPtr objc.Object
			options := map[appkit.WorkspaceDesktopImageOptionKey]objc.IObject{}
			success := workspace.SetDesktopImageURLForScreenOptionsError(url, screen, options, unsafe.Pointer(&errPtr))

			if !success {
				setError = fmt.Errorf("failed to set desktop image")
			}
		}
	})

	return setError
}
