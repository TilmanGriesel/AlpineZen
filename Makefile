APP_NAME := alpinzen-wallpaper
BUILD_DIR := build
DIST_DIR := dist
PLATFORMS := windows linux darwin
ARCHS := amd64 arm64

# Platform detection helpers
CURRENT_PLATFORM := $(shell uname -s | tr '[:upper:]' '[:lower:]' | sed -e 's/mingw64_nt.*/windows/' -e 's/msys_nt.*/windows/')
CURRENT_ARCH := $(shell uname -m | sed -e 's/arm64/arm64/' -e 's/x86_64/amd64/')

# Define extension function - returns .exe for Windows, empty for others
EXT = $(if $(filter windows,$1),.exe,)

.PHONY: all clean package release help build-run build-ui generate-windows-resources generate-windows-icon

all: build-run

clean:
	rm -rf $(BUILD_DIR) $(DIST_DIR)

# Windows-specific resource generation
generate-windows-resources:
	./tool/generate_windows_resources.sh

generate-windows-icon:
	mkdir -p assets/windows
	if command -v magick >/dev/null 2>&1; then \
		echo "Generating Windows icon from macOS icons..."; \
		magick assets/macos/AppIcon/AppIconSunrise.iconset/icon_512x512.png -define icon:auto-resize=16,32,48,64,128,256 assets/windows/MenuBarIcon.ico; \
	else \
		echo "ImageMagick not found. Windows icon generation skipped."; \
	fi

# Build target for a specific platform-arch combination
build-%:
	$(eval parts := $(subst -, ,$*))
	$(eval platform := $(word 1,$(parts)))
	$(eval arch := $(word 2,$(parts)))
	$(eval ext := $(call EXT,$(platform)))
	mkdir -p $(BUILD_DIR) $(DIST_DIR)/$(platform)-$(arch)
	if [ "$(platform)" = "windows" ]; then \
		$(MAKE) generate-windows-resources; \
		$(MAKE) generate-windows-icon; \
	fi
	if [ "$(platform)" = "darwin" ] || [ "$(platform)" = "windows" ]; then \
		CGO_ENABLED=1 GOOS=$(platform) GOARCH=$(arch) \
		go build -o $(BUILD_DIR)/AlpineZenHelper$(ext) ./cmd/cli; \
		CGO_ENABLED=1 GOOS=$(platform) GOARCH=$(arch) \
		go build -o $(BUILD_DIR)/$(APP_NAME)-$(platform)-$(arch)$(ext) ./cmd/ui; \
	else \
		GOOS=$(platform) GOARCH=$(arch) \
		go build -o $(BUILD_DIR)/$(APP_NAME)-$(platform)-$(arch)$(ext) ./cmd/cli; \
	fi
	mv $(BUILD_DIR)/* $(DIST_DIR)/$(platform)-$(arch)/
	if [ "$(platform)" = "windows" ] && [ -f "assets/windows/MenuBarIcon.ico" ]; then \
		cp assets/windows/MenuBarIcon.ico $(DIST_DIR)/$(platform)-$(arch)/; \
		echo "Windows icon copied to distribution directory."; \
	fi

# Current platform build targets
build-ui:
	mkdir -p $(BUILD_DIR)
	mkdir -p $(DIST_DIR)/$(CURRENT_PLATFORM)-$(CURRENT_ARCH)
	$(eval ext := $(call EXT,$(CURRENT_PLATFORM)))
	if [ "$(CURRENT_PLATFORM)" = "windows" ]; then \
		$(MAKE) generate-windows-resources; \
		$(MAKE) generate-windows-icon; \
	fi
	CGO_ENABLED=1 GOOS=$(CURRENT_PLATFORM) GOARCH=$(CURRENT_ARCH) \
	go build -o $(BUILD_DIR)/$(APP_NAME)-ui$(ext) ./cmd/ui
	mv $(BUILD_DIR)/* $(DIST_DIR)/$(CURRENT_PLATFORM)-$(CURRENT_ARCH)/
	if [ "$(CURRENT_PLATFORM)" = "windows" ] && [ -f "assets/windows/MenuBarIcon.ico" ]; then \
		cp assets/windows/MenuBarIcon.ico $(DIST_DIR)/$(CURRENT_PLATFORM)-$(CURRENT_ARCH)/; \
		echo "Windows icon copied to distribution directory."; \
	fi
	echo "UI built at $(DIST_DIR)/$(CURRENT_PLATFORM)-$(CURRENT_ARCH)/$(APP_NAME)-ui$(ext)"

build-run:
	$(eval ext := $(call EXT,$(CURRENT_PLATFORM)))
	$(MAKE) build-$(CURRENT_PLATFORM)-$(CURRENT_ARCH)
	./$(DIST_DIR)/$(CURRENT_PLATFORM)-$(CURRENT_ARCH)/$(APP_NAME)-$(CURRENT_PLATFORM)-$(CURRENT_ARCH)$(ext)

# Package and release targets
package:
	for platform in $(PLATFORMS); do \
		for arch in $(ARCHS); do \
			$(MAKE) build-$$platform-$$arch; \
		done; \
	done
	cd $(DIST_DIR) && \
	for platform in $(PLATFORMS); do \
		for arch in $(ARCHS); do \
			tar -czf $$platform-$$arch.tar.gz $$platform-$$arch/; \
		done; \
	done

release: clean package

# Additional utility targets
licenses:
	python3 tool/add_license_headers.py

regen-assets:
	./assets/macos/scripts/compile_assets.sh

repo-cleanup:
	./tool/rc_repo_cleanup.sh