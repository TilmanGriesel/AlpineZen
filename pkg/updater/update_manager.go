// SPDX-FileCopyrightText: 2024 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later

package updater

import (
	"context"
	"fmt"
	"time"

	"github.com/TilmanGriesel/AlpineZen/pkg/logging"
	"github.com/TilmanGriesel/AlpineZen/pkg/wallpaper"
	"github.com/sirupsen/logrus"
)

const (
	minUpdateIntervalMinutes = 1
)

type UpdaterManager struct {
	WallpaperManager *wallpaper.WallpaperManager
	Config           UpdateManagerConfig
	Logger           *logrus.Logger
	CancelCtx        context.CancelFunc
}

type UpdateManagerConfig struct {
	UpdateIntervalMinutes int
	DisableClock          bool
}

func NewUpdaterManager(wallpaperManager *wallpaper.WallpaperManager, config UpdateManagerConfig) *UpdaterManager {
	return &UpdaterManager{
		WallpaperManager: wallpaperManager,
		Config:           config,
		Logger:           logging.GetLogger(),
	}
}

func (um *UpdaterManager) validateUpdateInterval(interval int) error {
	if interval < minUpdateIntervalMinutes {
		return fmt.Errorf("update interval is too short; minimum is %d minutes", minUpdateIntervalMinutes)
	}
	return nil
}

func (um *UpdaterManager) runClockUpdater(ctx context.Context, tickerDuration time.Duration) {
	calculateTimeUntilNextInterval := func() time.Duration {
		now := time.Now()
		elapsed := now.Sub(now.Truncate(time.Minute))
		return tickerDuration - (elapsed % tickerDuration)
	}

	adjustedTimeUntilNextInterval := calculateTimeUntilNextInterval()
	if adjustedTimeUntilNextInterval < 0 {
		adjustedTimeUntilNextInterval = 0
	}

	um.Logger.WithField("timeUntilNextUpdate", adjustedTimeUntilNextInterval).
		Debug("Clock-updater is syncing with next update interval")

	time.Sleep(adjustedTimeUntilNextInterval)
	um.Logger.Debug("Initial clock sync complete")

	for {
		select {
		case <-ctx.Done():
			um.Logger.Info("Clock updater stopped")
			return
		default:
			um.WallpaperManager.UpdateWallpaper(false, false)

			adjustedTimeUntilNextInterval := calculateTimeUntilNextInterval()
			if adjustedTimeUntilNextInterval < 0 {
				adjustedTimeUntilNextInterval = 0
			}
			um.Logger.WithField("timeUntilNextUpdate", adjustedTimeUntilNextInterval).
				Debug("Clock-updater is resyncing with next update interval")
			time.Sleep(adjustedTimeUntilNextInterval)
			um.Logger.Debug("Clock resync complete")
		}
	}
}

func (um *UpdaterManager) runFullUpdater(ctx context.Context, updateInterval time.Duration) {
	um.Logger.Info("Performing initial update")
	um.WallpaperManager.UpdateWallpaper(true, true)

	if !um.Config.DisableClock {
		go um.runClockUpdater(ctx, time.Minute)
	}

	calculateTimeUntilNextInterval := func() time.Duration {
		now := time.Now()
		elapsed := now.Sub(now.Truncate(time.Hour))
		positiveBufferTime := 30 * time.Second
		return updateInterval - (elapsed % updateInterval) + positiveBufferTime
	}

	adjustedTimeUntilNextInterval := calculateTimeUntilNextInterval()
	um.Logger.WithField("timeUntilNextUpdate", adjustedTimeUntilNextInterval).
		Debug("Full-updater is syncing with next update interval")

	time.Sleep(adjustedTimeUntilNextInterval)
	um.Logger.Debug("Initial updater sync complete")

	for {
		select {
		case <-ctx.Done():
			um.Logger.Info("Full updater stopped")
			return
		default:
			um.WallpaperManager.UpdateWallpaper(true, false)

			adjustedTimeUntilNextInterval := calculateTimeUntilNextInterval()
			um.Logger.WithField("timeUntilNextUpdate", adjustedTimeUntilNextInterval).
				Debug("Full-updater is resyncing with next update interval")
			time.Sleep(adjustedTimeUntilNextInterval)
			um.Logger.Debug("Updater resync complete")
		}
	}
}

func (um *UpdaterManager) StartUpdater() {
	updateIntervalMinutes := um.Config.UpdateIntervalMinutes

	if err := um.validateUpdateInterval(updateIntervalMinutes); err != nil {
		um.Logger.WithField("updateIntervalMinutes", updateIntervalMinutes).
			WithError(err).
			Fatal("Invalid update interval")
	}
	um.Logger.WithField("updateIntervalMinutes", updateIntervalMinutes).Info("Starting updater")

	ctx, cancelFunc := context.WithCancel(context.Background())
	um.CancelCtx = cancelFunc

	go um.runFullUpdater(ctx, time.Duration(updateIntervalMinutes)*time.Minute)
}

func (um *UpdaterManager) StopUpdater() {
	if um.CancelCtx != nil {
		um.CancelCtx()
		um.Logger.Info("Updater cancellation requested")
	} else {
		um.Logger.Warn("Updater is not running or has already been stopped")
	}
}
