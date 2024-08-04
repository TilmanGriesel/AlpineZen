#!/bin/bash

# SPDX-FileCopyrightText: 2024 Tilman Griesel
#
# SPDX-License-Identifier: GPL-3.0-or-later

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_ICON_DIR="$SCRIPT_DIR/../AppIcon"
ASSETS_DIR="$SCRIPT_DIR/../Assets.xcassets"
DIST_DIR="$SCRIPT_DIR/../dist"

mkdir -p "$DIST_DIR"

create_icns() {
    local iconset=$1
    local output=$2

    echo "Creating $output..."
    iconutil -c icns "$APP_ICON_DIR/$iconset.iconset" -o "$DIST_DIR/$output.icns" || {
        echo "Failed to create $output ICNS file"
        exit 1
    }
}

compile_assets() {
    echo "Compiling assets..."

    MINIMUM_DEPLOYMENT_TARGET="10.13"
    PLATFORM="macosx"

    if actool --output-format human-readable-text \
              --notices --warnings \
              --minimum-deployment-target "$MINIMUM_DEPLOYMENT_TARGET" \
              --platform "$PLATFORM" \
              --compile "$DIST_DIR" \
              "$ASSETS_DIR"; then

        if [ -f "$DIST_DIR/Assets.car" ]; then
            echo "Assets compiled successfully."
        else
            echo "Assets.car file not found in $DIST_DIR"
            exit 1
        fi

    else
        echo "Failed to compile assets"
        exit 1
    fi
}

create_icns "AppIconSunset" "AlpineZenSunset"
create_icns "AppIconSunrise" "AlpineZenSunrise"

compile_assets

echo "All assets created successfully."
