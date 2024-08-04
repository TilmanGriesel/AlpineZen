APP_NAME := alpinzen-wallpaper
BUILD_DIR := build
DIST_DIR := dist

PLATFORMS := windows linux darwin
ARCHS := amd64 arm64
EXECUTABLES := $(foreach platform,$(PLATFORMS),$(foreach arch,$(ARCHS),$(BUILD_DIR)/$(APP_NAME)-$(platform)-$(arch)))

.PHONY: all clean package release help build-run

all: build-run

clean:
	rm -rf $(BUILD_DIR) $(DIST_DIR)

# Pattern rule to build for each platform and architecture
$(BUILD_DIR)/$(APP_NAME)-%:
	@platform=$(word 1,$(subst -, ,$*)); \
	arch=$(word 2,$(subst -, ,$*)); \
	mkdir -p $(BUILD_DIR); \
	mkdir -p $(DIST_DIR)/$$platform-$$arch; \
	if [ "$$platform" = "darwin" ]; then \
	  CGO_ENABLED=1 GOOS=$$platform GOARCH=$$arch \
	  go build -o $(BUILD_DIR)/AlpineZenHelper ./cmd/cli; \
	  CGO_ENABLED=1 GOOS=$$platform GOARCH=$$arch \
	  go build -o $(BUILD_DIR)/$(APP_NAME)-$$platform-$$arch ./cmd/ui; \
	else \
	  GOOS=$$platform GOARCH=$$arch \
	  go build -o $(BUILD_DIR)/$(APP_NAME)-$$platform-$$arch$(if $(findstring windows,$$platform),.exe,) ./cmd/cli; \
	fi; \
	mv $(BUILD_DIR)/* $(DIST_DIR)/$$platform-$$arch/

build-%:
	$(MAKE) $(BUILD_DIR)/$(APP_NAME)-$*

build-run:
	@os=$(shell uname -s | tr '[:upper:]' '[:lower:]'); \
	arch=$(shell uname -m | sed -e 's/arm64/arm64/' -e 's/x86_64/amd64/'); \
	$(MAKE) build-$$os-$$arch; \
	./$(DIST_DIR)/$$os-$$arch/$(APP_NAME)-$$os-$$arch$(if $(findstring windows,$$os),.exe,)

package: $(EXECUTABLES)
	@cd $(DIST_DIR) && \
	for platform in $(PLATFORMS); do \
	  for arch in $(ARCHS); do \
	    tar -czf $$platform-$$arch.tar.gz $$platform-$$arch/; \
	  done \
	done

release: clean package

# Helper targets
licenses:
	python3 tools/add_license_headers.py

regen-assets:
	./assets/macos/scripts/compile_assets.sh

repo-cleanup:
	bash tools/rc_repo_cleanup.sh
