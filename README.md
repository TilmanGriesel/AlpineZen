<p align="center"><img src="https://raw.githubusercontent.com/TilmanGriesel/AlpineZen/8c27063d52c33f1848c552c5df1f9d2000e73da7/docs/public/assets/brand/alpinezen_banner_eggshell.svg"/><br/></p>

[![Gosec](https://github.com/TilmanGriesel/AlpineZen/actions/workflows/gosec.yml/badge.svg)](https://github.com/TilmanGriesel/AlpineZen/actions/workflows/gosec.yml)
[![VulnCheck](https://github.com/TilmanGriesel/AlpineZen/actions/workflows/vulncheck.yml/badge.svg)](https://github.com/TilmanGriesel/AlpineZen/actions/workflows/vulncheck.yml)
[![Test Job](https://github.com/TilmanGriesel/AlpineZen/actions/workflows/test.yml/badge.svg)](https://github.com/TilmanGriesel/AlpineZen/actions/workflows/test.yml)
[![Tag Release Version](https://github.com/TilmanGriesel/AlpineZen/actions/workflows/versiontag.yml/badge.svg)](https://github.com/TilmanGriesel/AlpineZen/actions/workflows/versiontag.yml)
[![Deploy docs site to Pages](https://github.com/TilmanGriesel/AlpineZen/actions/workflows/docs-deploy.yml/badge.svg)](https://github.com/TilmanGriesel/AlpineZen/actions/workflows/docs-deploy.yml)

---

**AlpineZen Wallpaper** is an open-source dynamic wallpaper utility that enhances your workspace by setting wallpapers that update periodically. It specializes in integrating live webcam images, creating setups that reflect natural rhythms of day and bringing the outdoors into your work environment.

## Features

- Dynamic wallpaper updates from images sources like webcams
- Configurable update intervals
- Optional clock overlay with customization
- Image processing capabilities (contrast, saturation, brightness, etc.)
- Cross-platform compatibility (Windows, Linux, macOS)
- Headless mode for server environments

## Installation

### Prerequisites

- Go 1.22.3 or later
- Make (for building from source)

### Building from Source

1. Clone the repository:
```bash
git clone https://github.com/TilmanGriesel/AlpineZen.git
cd AlpineZen
```

2. Build the project:
```bash
make
```

The built executables will be available in the `dist` directory.

## Usage

### Basic Usage

```bash
alpinezen --name fellhorn --type default
```

### Configuration Options

#### Core Configuration
- `--name, -n`: Name of configuration profile (default: "fellhorn")
- `--type, -t`: Type of configuration profile (default: "default")
- `--config-path`: Path to custom configuration file
- `--config-repository`: URL of remote configuration repository

#### Wallpaper Settings
- `--wallpaper-width`: Width of wallpaper in pixels (default: 3840)
- `--wallpaper-height`: Height of wallpaper in pixels (default: 2160)

#### Clock Overlay Settings
- `--clock-disable`: Disable clock overlay
- `--clock-time-format`: Set time format (default: "15:04")
- `--clock-font-path`: Custom font file path
- `--clock-font-size`: Font size in points (default: 112)
- `--clock-font-dpi`: Font DPI for crisp rendering (default: 144)
- `--clock-font-opacity-min`: Minimum font opacity (default: 0.2)
- `--clock-font-opacity-max`: Maximum font opacity (default: 0.92)
- `--clock-font-color`: Font color in hex code (default: "#FFFFFF")
- `--clock-horizontal-offset`: Horizontal text offset
- `--clock-vertical-offset`: Vertical text offset

#### Runtime Options
- `--prepare`: Setup folder structure and exit
- `--version-show`: Display version information
- `--runtime-headless`: Enable headless mode for server environments
- `--runtime-cpu-cores`: Number of CPU cores to use
- `--loglevel`: Logging verbosity (0=Warn, 1=Info, 2=Debug, 3=Trace)

### Example Commands

1. Basic setup with default configuration:
```bash
alpinezen
```

2. Custom wallpaper dimensions with clock disabled:
```bash
alpinezen --wallpaper-width 1920 --wallpaper-height 1080 --clock-disable
```

3. Custom clock styling:
```bash
alpinezen --clock-font-color "#FF5733" --clock-font-size 96 --clock-time-format "15:04:05"
```

4. Headless mode for servers:
```bash
alpinezen --runtime-headless --loglevel 2
```

## Configuration File

AlpineZen uses YAML configuration files. Here's a sample structure:

```yaml
input:
  url: "https://example.com/webcam.jpg"
  crop_factor: 1.0
  offset_x: 0.0
  offset_y: 0.0

image_processing:
  contrast: 1.0
  saturation: 1.0
  brightness: 0.0
  hue: 0.0
  gamma: 1.0
  black_point: 0.0
  white_point: 1.0
  shadow_strength: 1.0
  blur_strength: 0.0
  sharpen_strength: 0.0
  max_noise_opacity: 0.0
  noise_scale: 1

scheduling:
  update_interval_minutes: 5

output:
  blend: false
  save_path: ""
```

## Application Directory Structure

The application maintains its files in the following structure:

```
~/.alpinezen_wallpaper/
├── config/
│   └── gui.yaml
├── files/
│   └── [hash]/
│       ├── .tmp/
│       │   ├── image
│       │   └── cache.png
│       └── proc/
│           └── [hash].png
├── latest.png
├── log/
│   └── alpinezen_cli.log
└── repos/
    └── AlpineZen-Basecamp-main/
```

## Error Handling

The application provides detailed logging with different verbosity levels. Logs are stored in:
```
~/.alpinezen_wallpaper/log/alpinezen_cli.log
```

## Contributing

Contributions are welcome! Please ensure your code follows these guidelines:
- Include appropriate license headers (GPL-3.0-or-later)
- Add tests for new functionality
- Follow Go best practices and coding standards
- Document new features and changes

## Roadmap
- [x] **1.0.x** Release initial version with macOS tray UI
- [ ] **1.1.x** Windows tray UI
- [ ] **1.2.x** Linux pre and post update hooks

Long term goals:
- Proper mixed aspect ratio multi-monitor support

## Acknowledgments

- Image processing: https://github.com/disintegration/imaging
- Native Mac APIs for Go https://github.com/progrium/darwinkit
- Logging: https://github.com/sirupsen/logrus
- CLI framework:  https://github.com/spf13/cobra
- Font: Lora (embedded) https://fonts.google.com/specimen/Lora
