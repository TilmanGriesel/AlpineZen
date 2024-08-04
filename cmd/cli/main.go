// SPDX-FileCopyrightText: 2025 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later AND LicenseRef-AlpineZen-Trademark

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/TilmanGriesel/AlpineZen/pkg/logging"
	"github.com/TilmanGriesel/AlpineZen/pkg/postprocess/render"
	"github.com/TilmanGriesel/AlpineZen/pkg/repository"
	"github.com/TilmanGriesel/AlpineZen/pkg/updater"
	"github.com/TilmanGriesel/AlpineZen/pkg/util"
	"github.com/TilmanGriesel/AlpineZen/pkg/wallpaper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	version     = "[dev]"
	buildNumber = "[0]"
	logger      = logging.GetLogger()
)

type Config struct {
	// Primary Configuration
	Name       string
	Type       string
	Path       string
	Repository string

	// Wallpaper Configuration
	Width  int
	Height int

	// Clock Configuration
	DisableClock       bool
	TimeFormat         string
	FontPath           string
	FontSize           float64
	FontDPI            float64
	FontOpacityMin     float64
	FontOpacityMax     float64
	FontColorHex       string
	DisableOSWallpaper bool

	// Clock Position Offsets
	ClockPositionConfig render.FontPositionConfig

	// Advanced Configuration
	PrepareOnly bool
	ShowVersion bool
	Headless    bool
	NumCores    int
	LogLevel    int
}

type Application struct {
	Config           Config
	UpdaterCancelCtx context.CancelFunc
	UpdaterManager   *updater.UpdaterManager
	WallpaperManager *wallpaper.WallpaperManager
}

func displayBanner() {
	fmt.Println(`
        _   _      _          ____
       /_\ | |_ __(_)_ _  ___|_  /___ _ _
      / _ \| | '_ \ | ' \/ -_)/ // -_) ' \
     /_/ \_\_| .__/_|_||_\___/___\___|_||_|
             |_|`)
	fmt.Printf("%10sAlpineZen CLI %s.%s\n\n", "", version, buildNumber)
}

func (app *Application) setupDefaultRepository(appDirPath string) error {
	if app.Config.Path != "" {
		return nil
	}

	localRepoPath, err := util.GetDefaultRepoPath()
	if err != nil {
		return fmt.Errorf("failed to get default repository path: %v", err)
	}
	if err := repository.DownloadAndExtractZip(app.Config.Repository, localRepoPath); err != nil {
		return fmt.Errorf("failed to download and extract default repository: %v", err)
	}

	logger.Info("Default repository updated")
	repoFolderName := repository.GetRepoFolderName(app.Config.Repository)
	app.Config.Path = filepath.Join(localRepoPath, repoFolderName, app.Config.Name, app.Config.Type+".yaml")

	return nil
}

func (app *Application) initialize(cmd *cobra.Command, args []string) error {
	appDirPath, err := util.GetAppDirPath()
	if err != nil {
		return fmt.Errorf("failed to get application directory path: %v", err)
	}

	switch app.Config.LogLevel {
	case 0:
		logger.SetLevel(logrus.WarnLevel)
	case 1:
		logger.SetLevel(logrus.InfoLevel)
	case 2:
		logger.SetLevel(logrus.DebugLevel)
	default:
		logger.SetLevel(logrus.TraceLevel)
	}

	if app.Config.ShowVersion {
		versionInfo := logrus.Fields{
			"version":         version,
			"build_number":    buildNumber,
			"os":              runtime.GOOS,
			"arch":            runtime.GOARCH,
			"core_count":      runtime.NumCPU(),
			"runtime_version": runtime.Version(),
		}

		logger.WithFields(versionInfo).Info("Version information")
		return nil
	}

	if app.Config.NumCores > 0 && app.Config.NumCores <= runtime.NumCPU() {
		runtime.GOMAXPROCS(app.Config.NumCores)
	} else {
		runtime.GOMAXPROCS(runtime.NumCPU())
		logger.WithFields(logrus.Fields{
			"available_cores": runtime.NumCPU(),
			"requested_cores": app.Config.NumCores,
		}).Warn("Invalid number of cores specified. Using maximum available cores")
	}

	displayBanner()

	if err := app.setupDefaultRepository(appDirPath); err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	if app.Config.PrepareOnly {
		logger.Info("Prepare only complete. Exiting.")
		os.Exit(0)
	}

	colorRGBA, err := util.ParseHexColor(app.Config.FontColorHex)
	if err != nil {
		logger.WithError(err).Fatal("Invalid font color!")
	}

	app.WallpaperManager, err = wallpaper.NewWallpaperManager(app.Config.Path)
	if err != nil {
		logger.WithError(err).Fatal("Unable to instantiate wallpaper manager!")
	}

	updateManagerConfig := updater.UpdateManagerConfig{
		UpdateIntervalMinutes: app.WallpaperManager.WallpaperManagerConfig.Scheduling.UpdateIntervalMinutes,
		DisableClock:          app.Config.DisableClock,
	}

	app.UpdaterManager = updater.NewUpdaterManager(app.WallpaperManager, updateManagerConfig)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Setup wallpaper configuration
	app.WallpaperManager.WallpaperConfig.DisableClock = app.Config.DisableClock
	app.WallpaperManager.WallpaperConfig.DisableOSWallpaperUpdate = app.Config.Headless
	app.WallpaperManager.WallpaperConfig.TargetDimensions = wallpaper.Dimensions{
		Width:  app.Config.Width,
		Height: app.Config.Height,
	}

	// Setup font configuration for clock
	app.WallpaperManager.WallpaperConfig.FontConfigClock = render.FontConfig{
		FontPath:   app.Config.FontPath,
		Size:       app.Config.FontSize,
		DPI:        app.Config.FontDPI,
		Color:      colorRGBA,
		MinOpacity: app.Config.FontOpacityMin,
		MaxOpacity: app.Config.FontOpacityMax,
		TimeFormat: app.Config.TimeFormat,

		Position: render.FontPositionConfig{
			HorizontalAlignment:    render.AlignCenter,
			VerticalAlignment:      render.AlignMiddle,
			PaddingTop:             0,
			PaddingBottom:          0,
			PaddingLeft:            0,
			PaddingRight:           0,
			HorizontalCenterOffset: app.Config.ClockPositionConfig.HorizontalCenterOffset,
			VerticalCenterOffset:   app.Config.ClockPositionConfig.VerticalCenterOffset,
		},
	}

	app.UpdaterManager.StartUpdater()

	return nil
}

func (app *Application) processFlags() {
	// Usage
	var rootCmd = &cobra.Command{
		Use:   "alpinezen",
		Short: "AlpineZen - Dynamic wallpaper utility",
		Long:  `AlpineZen is an open-source tool that enhances your workspace by setting dynamic wallpapers that update periodically. By integrating live webcam images, it creates setups that reflect natural rhythms of day, bringing outdoors into your work environment.`,
		RunE:  app.initialize,
	}

	// Config flags
	rootCmd.Flags().StringVarP(&app.Config.Name, "name", "n", "fellhorn",
		"Name of configuration profile to use.")
	rootCmd.Flags().StringVarP(&app.Config.Type, "type", "t", "default",
		"Type of configuration profile (e.g., 'default', 'blur').")

	rootCmd.Flags().StringVar(&app.Config.Path, "config-path", "",
		"Path to a custom configuration file. If omitted, configuration will be retrieved from default repository.")
	rootCmd.Flags().StringVar(&app.Config.Repository, "config-repository", "https://github.com/TilmanGriesel/AlpineZen-Basecamp/archive/refs/heads/main.zip",
		"URL of remote configuration repository to fetch profiles from.")

	// Wallpaper flags
	rootCmd.Flags().IntVar(&app.Config.Width, "wallpaper-width", 3840,
		"Width of wallpaper in pixels.")
	rootCmd.Flags().IntVar(&app.Config.Height, "wallpaper-height", 2160,
		"Height of wallpaper in pixels.")

	// Clock flags
	rootCmd.Flags().BoolVar(&app.Config.DisableClock, "clock-disable", false,
		"Disable clock overlay on wallpaper.")
	rootCmd.Flags().StringVar(&app.Config.TimeFormat, "clock-time-format", "15:04",
		"Set time format for clock overlay on wallpaper. Use '15:04' for 24-hour format or customize as needed.")
	rootCmd.Flags().StringVar(&app.Config.FontPath, "clock-font-path", "",
		"Path to font file for clock overlay.")
	rootCmd.Flags().Float64Var(&app.Config.FontSize, "clock-font-size", 112,
		"Font size for clock overlay in points.")
	rootCmd.Flags().Float64Var(&app.Config.FontDPI, "clock-font-dpi", 144,
		"DPI for clock font to ensure crisp rendering.")
	rootCmd.Flags().Float64Var(&app.Config.FontOpacityMin, "clock-font-opacity-min", 0.2,
		"Minimum opacity for clock font (0.0 - transparent, 1.0 - opaque).")
	rootCmd.Flags().Float64Var(&app.Config.FontOpacityMax, "clock-font-opacity-max", 0.92,
		"Maximum opacity for clock font (0.0 - transparent, 1.0 - opaque).")
	rootCmd.Flags().StringVar(&app.Config.FontColorHex, "clock-font-color", "#FFFFFF",
		"Clock font color in hex code (e.g., '#FF5733' for orange).")
	rootCmd.Flags().IntVar(&app.Config.ClockPositionConfig.HorizontalCenterOffset, "clock-horizontal-offset", 0,
		"Horizontal offset for clock text when centered.")
	rootCmd.Flags().IntVar(&app.Config.ClockPositionConfig.VerticalCenterOffset, "clock-vertical-offset", 0,
		"Vertical offset for clock text when centered.")

	// Runtime flags
	rootCmd.Flags().BoolVar(&app.Config.PrepareOnly, "prepare", false,
		"Setup folder structure, pull default repo and exit.")
	rootCmd.Flags().BoolVar(&app.Config.ShowVersion, "version-show", false,
		"Display version information and exit.")
	rootCmd.Flags().BoolVar(&app.Config.Headless, "runtime-headless", false,
		"Enable headless mode to prevent interaction with OS wallpaper (useful for server environments).")
	rootCmd.Flags().IntVar(&app.Config.NumCores, "runtime-cpu-cores", runtime.NumCPU()/2,
		"Number of CPU cores to use. If exceeds available cores, all cores will be used.")
	rootCmd.Flags().IntVar(&app.Config.LogLevel, "loglevel", 1,
		"Logging verbosity level: 0 = Warn, 1 = Info, 2 = Debug, 3 = Trace.")

	if err := rootCmd.Execute(); err != nil {
		logger.Fatalf("Application error: %v", err)
	}
}

func main() {
	logging.SetLogFileName("alpinezen_cli.log")

	app := &Application{}
	app.processFlags()

	// Handle cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for OS signals in a separate goroutine
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case sig := <-sigChan:
				if sig == syscall.SIGINT || sig == syscall.SIGTERM {
					logger.WithField("signal", sig).Info("Received termination signal. Initiating shutdown")
					cancel()
					app.UpdaterManager.StopUpdater()
					logger.Info("Graceful shutdown complete")
					os.Exit(0)
				}
			case <-ctx.Done():
				logger.Info("Context canceled, exiting signal goroutine")
				return
			}
		}
	}()

	// Main event loop
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			logger.Info("Context canceled, exiting")
			return
		}
	}
}
