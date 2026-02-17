# ── Configuration ────────────────────────────────────────────────────────────

# Local arch by default — no cross-compilation needed for dev
GOOS   ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

EXT_DIR := packages/vscode

# All supported platforms (for package-all / CI release only)
# Format: VSCE_TARGET:GOOS:GOARCH
PLATFORMS := \
	darwin-arm64:darwin:arm64 \
	darwin-x64:darwin:amd64 \
	linux-x64:linux:amd64 \
	linux-arm64:linux:arm64 \
	win32-x64:windows:amd64

# ── Dev shortcuts ────────────────────────────────────────────────────────────
# These build for the local platform only. Use *-all variants for cross-platform.

.PHONY: build publish clean

## Build everything for the local platform
build: build-lsp build-visualizer build-skills build-extension

## Publish all .vsix files in packages/vscode/ to registries
publish: publish-vscode publish-ovsx

# ── Build targets ────────────────────────────────────────────────────────────

.PHONY: build-lsp build-visualizer build-skills build-extension

## Build the twf binary for the current (or specified) platform
build-lsp:
	@mkdir -p $(EXT_DIR)/bin
	cd tools/lsp && \
		GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 \
		go build -o ../../$(EXT_DIR)/bin/twf$(if $(filter windows,$(GOOS)),.exe) ./cmd/twf
	@echo "Built twf for $(GOOS)/$(GOARCH)"

## Build the visualizer webview into the extension
build-visualizer:
	cd tools/visualizer && npm run build:webview
	@echo "Built visualizer"

## Copy skills into the extension package
build-skills:
	@mkdir -p $(EXT_DIR)/skills
	rsync -a --delete skills/ $(EXT_DIR)/skills/
	@echo "Copied skills"

## Compile the extension TypeScript
build-extension: build-skills
	cd $(EXT_DIR) && npm run compile
	@echo "Compiled extension"

# ── Test targets ─────────────────────────────────────────────────────────────

.PHONY: test vet

## Run Go tests
test:
	cd tools/lsp && go test ./...

## Run Go vet
vet:
	cd tools/lsp && go vet ./...

# ── Package targets ──────────────────────────────────────────────────────────

.PHONY: package package-platform package-all

## Package a VSIX for the local platform
package: build
	cd $(EXT_DIR) && npx @vscode/vsce package
	@echo "Packaged VSIX"

## Package a VSIX for a single explicit target (used by CI matrix)
## Usage: make package-platform VSCE_TARGET=darwin-arm64 GOOS=darwin GOARCH=arm64
package-platform: build-lsp
	cd $(EXT_DIR) && npx @vscode/vsce package --target $(VSCE_TARGET)
	@echo "Packaged VSIX for $(VSCE_TARGET)"

## Package VSIXes for all platforms
package-all: build-visualizer build-skills build-extension
	@for entry in $(PLATFORMS); do \
		target=$$(echo $$entry | cut -d: -f1); \
		os=$$(echo $$entry | cut -d: -f2); \
		arch=$$(echo $$entry | cut -d: -f3); \
		echo "Packaging $$target ($$os/$$arch)..."; \
		$(MAKE) package-platform VSCE_TARGET=$$target GOOS=$$os GOARCH=$$arch; \
	done
	@echo "All platform packages built"

# ── Publish targets ──────────────────────────────────────────────────────────

.PHONY: publish-vscode publish-ovsx

## Publish all platform VSIXes to VS Code Marketplace
publish-vscode:
	@if [ -z "$(VSCE_TOKEN)" ]; then \
		echo "Error: VSCE_TOKEN not set"; exit 1; \
	fi
	@for vsix in $(EXT_DIR)/*.vsix; do \
		echo "Publishing $$vsix to VS Code Marketplace..."; \
		cd $(EXT_DIR) && npx @vscode/vsce publish --packagePath $$(basename $$vsix) -p $(VSCE_TOKEN) && cd ../..; \
	done

## Publish all platform VSIXes to Open VSX
publish-ovsx:
	@if [ -z "$(OVSX_TOKEN)" ]; then \
		echo "Error: OVSX_TOKEN not set"; exit 1; \
	fi
	@for vsix in $(EXT_DIR)/*.vsix; do \
		echo "Publishing $$vsix to Open VSX..."; \
		npx ovsx publish $$vsix -p $(OVSX_TOKEN); \
	done

# ── Release targets ──────────────────────────────────────────────────────────
# Bump version, commit, tag, and push — triggers the release workflow.
#   make release TYPE=patch        (auto-bump from latest tag)
#   make release TYPE=minor
#   make release TYPE=major
#   make release VERSION=1.2.3     (explicit version)

.PHONY: release

release:
	$(eval NEW_VERSION := $(shell bash scripts/version.sh "$(VERSION)" "$(TYPE)"))
	@if [ -z "$(NEW_VERSION)" ]; then exit 1; fi
	@echo "Releasing v$(NEW_VERSION)"
	@sed -i.bak 's/"version": *"[^"]*"/"version": "$(NEW_VERSION)"/' $(EXT_DIR)/package.json && rm -f $(EXT_DIR)/package.json.bak
	git add $(EXT_DIR)/package.json
	git commit -m "release: v$(NEW_VERSION)"
	git tag "v$(NEW_VERSION)"
	git push origin HEAD "v$(NEW_VERSION)"
	@echo "Pushed v$(NEW_VERSION) — release workflow will build and publish"

# ── Clean ────────────────────────────────────────────────────────────────────

.PHONY: clean

## Remove all build artifacts
clean:
	rm -rf $(EXT_DIR)/bin $(EXT_DIR)/dist $(EXT_DIR)/out $(EXT_DIR)/skills $(EXT_DIR)/*.vsix
	@echo "Cleaned"
