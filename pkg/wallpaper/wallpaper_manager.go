// SPDX-FileCopyrightText: 2024 Tilman Griesel
//
// SPDX-License-Identifier: GPL-3.0-or-later

package wallpaper

import (
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/TilmanGriesel/AlpineZen/pkg/logging"
	"github.com/TilmanGriesel/AlpineZen/pkg/postprocess"
	"github.com/TilmanGriesel/AlpineZen/pkg/postprocess/adjustment"
	"github.com/TilmanGriesel/AlpineZen/pkg/postprocess/render"
	"github.com/TilmanGriesel/AlpineZen/pkg/sanitizer"
	"github.com/sirupsen/logrus"

	"github.com/TilmanGriesel/AlpineZen/pkg/repository"
	"github.com/TilmanGriesel/AlpineZen/pkg/util"

	"github.com/disintegration/imaging"
	"gopkg.in/yaml.v2"
)

const (
	FileType           = ".png"
	MaxWatermarkHeight = 50
)

var (
	logger = logging.GetLogger()
)

type WallpaperManager struct {
	WallpaperManagerConfig WallpaperManagerConfig
	WallpaperConfig        WallpaperConfig
	noiseImg               image.NRGBA
	configPath             string
	updateCount            int
}

type WallpaperConfig struct {
	DisableClock             bool
	DisableOSWallpaperUpdate bool
	TargetDimensions         Dimensions
	FontConfigClock          render.FontConfig
}

type Dimensions struct {
	Width  int
	Height int
}

type WallpaperManagerConfig struct {
	Input struct {
		URL        string  `yaml:"url"`
		CropFactor float64 `yaml:"crop_factor"`
		OffsetX    float64 `yaml:"offset_x"`
		OffsetY    float64 `yaml:"offset_y"`
	} `yaml:"input"`
	ImageProcessing struct {
		Contrast        float64 `yaml:"contrast"`
		Saturation      float64 `yaml:"saturation"`
		Brightness      float64 `yaml:"brightness"`
		Hue             float64 `yaml:"hue"`
		Gamma           float64 `yaml:"gamma"`
		BlackPoint      float64 `yaml:"black_point"`
		WhitePoint      float64 `yaml:"white_point"`
		ShadowStrength  float64 `yaml:"shadow_strength"`
		BlurStrength    float64 `yaml:"blur_strength"`
		SharpenStrength float64 `yaml:"sharpen_strength"`
		MaxNoiseOpacity float64 `yaml:"max_noise_opacity"`
		NoiseScale      int     `yaml:"noise_scale"`
	} `yaml:"image_processing"`
	Scheduling struct {
		UpdateIntervalMinutes int `yaml:"update_interval_minutes"`
	} `yaml:"scheduling"`
	Output struct {
		Blend    bool   `yaml:"blend"`
		SavePath string `yaml:"save_path"`
	} `yaml:"output"`
}

func NewWallpaperManager(configPath string) (*WallpaperManager, error) {
	updater := &WallpaperManager{
		configPath: configPath,
	}

	if err := updater.LoadConfig(configPath); err != nil {
		return nil, err
	}

	return updater, nil
}

func (wm *WallpaperManager) LoadConfig(path string) error {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		logger.WithError(err).WithField("path", path).Error("Failed to open config file")
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&wm.WallpaperManagerConfig); err != nil {
		logger.WithError(err).Error("Failed to decode config file")
		return err
	}

	logger.WithField("path", path).Debug("Configuration loaded successfully")
	return nil
}

func (wm *WallpaperManager) setWallpaper(filepath string) error {
	logger.WithField("filepath", filepath).Debug("Set wallpaper from file")
	return SetWallpaper(filepath)
}

func (wm *WallpaperManager) enhanceImage(img image.Image) image.Image {
	processor := adjustment.NewImageEnhancer()
	processor.Contrast = wm.WallpaperManagerConfig.ImageProcessing.Contrast
	processor.Saturation = wm.WallpaperManagerConfig.ImageProcessing.Saturation
	processor.Brightness = wm.WallpaperManagerConfig.ImageProcessing.Brightness
	processor.Hue = wm.WallpaperManagerConfig.ImageProcessing.Hue
	processor.Gamma = wm.WallpaperManagerConfig.ImageProcessing.Gamma
	processor.BlackPoint = wm.WallpaperManagerConfig.ImageProcessing.BlackPoint
	processor.WhitePoint = wm.WallpaperManagerConfig.ImageProcessing.WhitePoint
	processor.ShadowStrength = wm.WallpaperManagerConfig.ImageProcessing.ShadowStrength

	enhancedImg := processor.ApplyEnhancements(img)
	sharpenedImage := imaging.Sharpen(enhancedImg, wm.WallpaperManagerConfig.ImageProcessing.SharpenStrength)
	return imaging.Resize(sharpenedImage, wm.WallpaperConfig.TargetDimensions.Width, wm.WallpaperConfig.TargetDimensions.Height, imaging.Lanczos)
}

func (wm *WallpaperManager) applyBlur(img image.Image) image.Image {
	return imaging.Blur(img, wm.WallpaperManagerConfig.ImageProcessing.BlurStrength)
}

func (wm *WallpaperManager) applyNoise(img image.Image, maxOpacity float64, scale int) image.Image {
	avgBrightness := postprocess.CalculateAverageBrightness(img)
	scaledOpacity := postprocess.CalculateScaledOpacity(avgBrightness, 0.0, 0.4, 0.01, maxOpacity)
	logger.WithField("maxNoiseOpacity", maxOpacity).WithField("scaledOpacity", scaledOpacity).WithField("averageBrightness", avgBrightness).Debug("Scaled noise opacity calculated")

	if avgBrightness > adjustment.NoiseBrightnessThreshold {
		logger.WithField("noiseOpacity", scaledOpacity).Debug("Applying noise")

		width := img.Bounds().Dx()
		height := img.Bounds().Dy()

		if scale < 1 {
			scale = 1
		}

		lowWidth := width / scale
		lowHeight := height / scale

		noiseImg := postprocess.CreateNoiseImage(lowWidth, lowHeight)
		noiseImg = imaging.Resize(noiseImg, width, height, imaging.NearestNeighbor)
		return imaging.Overlay(img, noiseImg, image.Pt(0, 0), scaledOpacity)
	}

	logger.Debug("Image is not bright enough, returning original image")
	return img
}

func (wm *WallpaperManager) cropImage(img image.Image, factor, offsetX, offsetY float64) image.Image {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	cropWidth := int(float64(width) / factor)
	cropHeight := int(float64(height) / factor)
	cropRect := image.Rect(
		(width-cropWidth)/2+int(offsetX*float64(width)),
		(height-cropHeight)/2+int(offsetY*float64(height)),
		(width+cropWidth)/2+int(offsetX*float64(width)),
		(height+cropHeight)/2+int(offsetY*float64(height)),
	)
	return imaging.Crop(img, cropRect)
}

func (wm *WallpaperManager) processImage(tempPath, finalImagePath string) (image.Image, error) {
	logger.WithField("tempPath", tempPath).WithField("finalImagePath", finalImagePath).Debug("Processing image")
	img, err := imaging.Open(tempPath)
	if err != nil {
		logger.WithError(err).Error("Failed to open image")
		return nil, err
	}

	finalImage := wm.cropImage(img, wm.WallpaperManagerConfig.Input.CropFactor, wm.WallpaperManagerConfig.Input.OffsetX, wm.WallpaperManagerConfig.Input.OffsetY)
	finalImage = wm.enhanceImage(finalImage)
	finalImage = wm.applyBlur(finalImage)
	finalImage = wm.applyNoise(finalImage, wm.WallpaperManagerConfig.ImageProcessing.MaxNoiseOpacity, wm.WallpaperManagerConfig.ImageProcessing.NoiseScale)

	watermarkPath := filepath.Join(filepath.Dir(wm.configPath), "watermark.png")
	watermark, err := imaging.Open(watermarkPath)
	if err == nil {
		if watermark.Bounds().Dy() > MaxWatermarkHeight {
			ratio := float64(MaxWatermarkHeight) / float64(watermark.Bounds().Dy())
			newWidth := int(float64(watermark.Bounds().Dx()) * ratio)
			watermark = imaging.Resize(watermark, newWidth, MaxWatermarkHeight, imaging.Lanczos)
		}

		offset := image.Pt(finalImage.Bounds().Dx()-watermark.Bounds().Dx()-20, finalImage.Bounds().Dy()-watermark.Bounds().Dy()-20)

		avgBrightness := postprocess.CalculateAverageBrightness(img)
		scaledOpacity := postprocess.CalculateScaledOpacity(avgBrightness, 0.0, 0.4, 0.2, 0.8)

		finalImage = imaging.Overlay(finalImage, watermark, offset, scaledOpacity)
	}

	return finalImage, nil
}

func (wm *WallpaperManager) cleanUpOldFiles(janitor *repository.Janitor, path string, deepClean bool) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logger.Debugf("Directory %s does not exist, skipping deep clean", path)
		return nil
	}

	if deepClean {
		if err := janitor.DeepClean(path); err != nil {
			logger.WithError(err).Fatalf("Deep clean failed for %s", path)
			return err
		}
	} else {
		if err := janitor.WipeThrough(path, 2); err != nil {
			logger.WithError(err).Fatalf("Wipe through failed for %s", path)
			return err
		}
	}

	return nil
}

func (wm *WallpaperManager) prepareDirectories(tempImageFilePath, imageFilePath string) error {
	if err := os.MkdirAll(filepath.Dir(tempImageFilePath), 0750); err != nil {
		logger.WithError(err).Error("Failed to create temp image directory")
		return err
	}

	if err := os.MkdirAll(filepath.Dir(imageFilePath), 0750); err != nil {
		logger.WithError(err).Error("Failed to create image directory")
		return err
	}

	return nil
}

func (wm *WallpaperManager) fetchAndProcessImage(tempImageFilePath, previousProcImageFilePath, imageFilePath string) (image.Image, error) {
	var finalImage image.Image
	var err error

	logger.WithField("tempImageFilePath", tempImageFilePath).WithField("imageFilePath", imageFilePath).Debug("Fetching new image from source")
	if err := util.DownloadImage(wm.WallpaperManagerConfig.Input.URL, tempImageFilePath, false); err != nil {
		logger.WithError(err).Warn("Failed to download image")
		return nil, err
	}

	logger.Debug("Sanitizing downloaded image")
	if err := sanitizer.SanitizeImage(tempImageFilePath); err != nil {
		logger.WithError(err).Fatal("Failed to sanitize image")
		return nil, err
	}

	logger.Debug("Processing new image")
	finalImage, err = wm.processImage(tempImageFilePath, imageFilePath)
	if err != nil {
		logger.WithError(err).Fatal("Failed to process image")
		return nil, err
	}

	if wm.WallpaperManagerConfig.Output.Blend && util.FileExists(previousProcImageFilePath) {
		previousImage, err := util.LoadImageFile(previousProcImageFilePath)
		if err != nil {
			logger.WithError(err).Warn("Failed to load previous processed image")
		} else {
			finalImage = imaging.Overlay(finalImage, previousImage, image.Pt(0, 0), 0.5)
		}
	}

	return finalImage, nil
}

func (wm *WallpaperManager) saveFinalImage(finalImage image.Image, imageFilePath, previousProcImageFilePath string, pngCompressionLevel imaging.EncodeOption) error {
	if err := imaging.Save(finalImage, previousProcImageFilePath, pngCompressionLevel); err != nil {
		logger.WithError(err).Fatal("Failed to save processed image")
		return err
	}

	if err := os.MkdirAll(filepath.Dir(imageFilePath), 0750); err != nil {
		logger.WithError(err).Fatal("Failed to create savePathFilePath directory")
		return err
	}

	if !wm.WallpaperConfig.DisableClock {
		finalImage = render.DrawTimeOnImage(finalImage, wm.WallpaperConfig.FontConfigClock)
	}

	if err := imaging.Save(finalImage, imageFilePath, pngCompressionLevel); err != nil {
		logger.WithError(err).Fatal("Failed to save final image")
		return err
	}

	return nil
}

func (wm *WallpaperManager) applyWallpaper(imageFilePath, latestFilePath string) error {
	// Set wallpaper if update is not disabled
	if !wm.WallpaperConfig.DisableOSWallpaperUpdate {
		if err := wm.setWallpaper(imageFilePath); err != nil {
			logger.WithError(err).Fatal("Failed to set wallpaper")
			return err
		}
		return nil
	}

	// Disabled OS wallpaper config is intended for headless server use
	// Serve the wallpaper as the original PNG and a JPG version as a convenient latest file.

	// Copy original image file to the latestFilePath
	if err := util.CopyFile(imageFilePath, latestFilePath); err != nil {
		logger.WithError(err).Warning("Failed to copy final image to latest")
		return err
	}

	// Open the copied file
	copiedFile, err := os.Open(filepath.Clean(latestFilePath))
	if err != nil {
		logger.WithError(err).Warning("Failed to open copied image file")
		return err
	}
	defer copiedFile.Close()

	// Decode the image
	img, _, err := image.Decode(copiedFile)
	if err != nil {
		logger.WithError(err).Warning("Failed to decode image file")
		return err
	}

	// Create a new file with .jpg extension
	jpgFilePath := latestFilePath
	if filepath.Ext(latestFilePath) != ".jpg" {
		jpgFilePath = latestFilePath[:len(latestFilePath)-len(filepath.Ext(latestFilePath))] + ".jpg"
	}

	jpgFile, err := os.Create(filepath.Clean(jpgFilePath))
	if err != nil {
		logger.WithError(err).Warning("Failed to create JPEG file")
		return err
	}
	defer jpgFile.Close()

	// Encode JPEG image
	opts := &jpeg.Options{Quality: 100}
	if err := jpeg.Encode(jpgFile, img, opts); err != nil {
		logger.WithError(err).Warning("Failed to encode image as JPEG")
		return err
	}

	// Set up archive folder path
	archiveFolderPath := filepath.Join(filepath.Dir(jpgFilePath), "archive")

	// Create archive folder
	if err := os.MkdirAll(archiveFolderPath, os.ModePerm); err != nil {
		logger.WithError(err).Warning("Failed to create archive directory")
		return err
	}

	// Copy JPEG file to archive folder
	archiveFilePath := filepath.Join(archiveFolderPath, filepath.Base(jpgFilePath))
	if err := util.CopyFile(jpgFilePath, archiveFilePath); err != nil {
		logger.WithError(err).Warning("Failed to archive JPEG file")
		return err
	}

	return nil
}

func (wm *WallpaperManager) UpdateWallpaper(fetchSource, deepClean bool) {
	logger.WithField("fetchSource", fetchSource).WithField("deepClean", deepClean).Debug("Updating wallpaper")

	if !fetchSource && deepClean {
		logger.Fatal("Deep clean requires source fetch")
		return
	}

	janitor := repository.NewJanitor(FileType)
	startTime := time.Now().UTC().UnixNano()
	timestampStr := strconv.FormatInt(startTime, 10)

	appDirPath, err := util.GetAppDirPath()
	if err != nil {
		logger.WithError(err).Error("Failed to get app directory path")
		return
	}

	wallpaperDirName := "files"
	hash := util.GenerateShortHash(wm.WallpaperManagerConfig.Input.URL, timestampStr)
	urlHash := util.GenerateShortHash(wm.WallpaperManagerConfig.Input.URL, "")

	wallpaperPath := filepath.Join(appDirPath, wallpaperDirName, urlHash)
	tempImagePath := filepath.Join(wallpaperPath, ".tmp")
	tempImageFilePath := filepath.Join(tempImagePath, "image")
	previousProcImageFilePath := filepath.Join(tempImagePath, "cache"+FileType)
	imagePath := filepath.Join(wallpaperPath, "proc")
	imageFilePath := filepath.Join(imagePath, hash+FileType)
	latestFilePath := filepath.Join(appDirPath, "latest"+FileType)

	err = wm.cleanUpOldFiles(janitor, imagePath, deepClean)
	if err != nil {
		logger.WithError(err).WithField("deepClean", deepClean).WithField("tempImagePath", tempImagePath).Fatal("Failed to clean up old files")
	}
	if deepClean {
		err = wm.cleanUpOldFiles(janitor, tempImagePath, true)

		if err != nil {
			logger.WithError(err).WithField("deepClean", deepClean).WithField("tempImagePath", tempImagePath).Fatal("Failed to clean up old files")
		}
	}

	if err := wm.prepareDirectories(tempImageFilePath, imageFilePath); err != nil {
		logger.WithError(err).Fatal("Failed to prepare directories")
	}

	var finalImage image.Image
	if fetchSource {
		finalImage, err = wm.fetchAndProcessImage(tempImageFilePath, previousProcImageFilePath, imageFilePath)
		if err != nil {
			logger.WithError(err).Error("Failed to fetch and process image")
			return
		}
	} else if util.FileExists(previousProcImageFilePath) {
		finalImage, err = util.LoadImageFile(previousProcImageFilePath)
		if err != nil {
			logger.WithError(err).Warn("Failed to load previous processed image")
			return
		}
	}

	pngCompressionLevel := imaging.PNGCompressionLevel(png.NoCompression)
	if err := wm.saveFinalImage(finalImage, imageFilePath, previousProcImageFilePath, pngCompressionLevel); err != nil {
		return
	}

	if err := wm.applyWallpaper(imageFilePath, latestFilePath); err != nil {
		return
	}

	wm.updateCount++
	timeSpend := time.Duration(time.Now().UTC().UnixNano() - startTime)

	logger.WithFields(logrus.Fields{
		"fetchSource": fetchSource,
		"deepClean":   deepClean,
		"updateCount": wm.updateCount,
		"timeSpend":   timeSpend.String(),
	}).Info("Wallpaper update completed")
}
