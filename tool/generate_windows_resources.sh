#!/bin/bash

echo "Generating Windows resources..."

# Create Windows icon from macOS icons
if command -v magick >/dev/null 2>&1; then
    echo "Generating Windows icon from macOS icons..."
    mkdir -p assets/windows
    magick assets/macos/AppIcon/AppIconSunrise.iconset/icon_512x512.png -define icon:auto-resize=16,32,48,64,128,256 assets/windows/MenuBarIcon.ico
fi

# Generate manifest resource
if command -v go-winres >/dev/null 2>&1; then
    go-winres make --in=cmd/ui/assets/app.manifest --out=cmd/ui/rsrc
elif command -v rsrc >/dev/null 2>&1; then
    rsrc -manifest cmd/ui/assets/app.manifest -o cmd/ui/rsrc.syso
else
    echo "Neither go-winres nor rsrc found. Installing rsrc..."
    go install github.com/akavel/rsrc@latest
    rsrc -manifest cmd/ui/assets/app.manifest -o cmd/ui/rsrc.syso
fi