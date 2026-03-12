NAME=sshpky
VERSION=$(shell cat VERSION)
BUILD_TIME=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(shell git rev-parse HEAD)
BUILD_NUMBER=$(shell date -u +'%Y%m%d%H%M%S')
TREE_STATE=$(shell if [ -z "$(shell git status --porcelain)" ]; then echo "clean"; else echo "dirty"; fi)
EXT_LD_FLAGS="-Wl"
LD_FLAGS="-s -w -X sshpky/pkg/common.version=$(VERSION) -X sshpky/pkg/common.buildTime=$(BUILD_TIME) -X sshpky/pkg/common.gitCommit=$(GIT_COMMIT) -X sshpky/pkg/common.buildNumber=$(BUILD_NUMBER) -X sshpky/pkg/common.treeState=$(TREE_STATE) -extldflags=$(EXT_LD_FLAGS)"
OUTPUT_DIR=./bin

clean:
	rm -rf _build/ release/ $(OUTPUT_DIR)/ dist/

build:
	@echo "Building "
	go mod tidy
	go build -tags release -ldflags $(LD_FLAGS) -o $(OUTPUT_DIR)/$(NAME)
	@bash ./scripts/completions.sh

build-dev:
	@echo "Building development version "
	mkdir -p $(OUTPUT_DIR)
	go build -ldflags "-w -X main.version=$(VERSION)-dev -X main.build=$(BUILD) -extldflags=$(EXT_LD_FLAGS)" -o $(OUTPUT_DIR)/$(NAME)-dev

build-all: clean
	@echo "Building for all platforms from $(SRC_DIR)..."
	mkdir -p _build $(OUTPUT_DIR)
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=1 go build -tags release -ldflags $(LD_FLAGS) -o _build/$(NAME)-$(VERSION)-darwin-arm64
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=1 go build -tags release -ldflags $(LD_FLAGS) -o _build/$(NAME)-$(VERSION)-darwin-amd64
	GOOS=linux   GOARCH=amd64 go build -tags release -ldflags $(LD_FLAGS) -o _build/$(NAME)-$(VERSION)-linux-amd64
	GOOS=linux   GOARCH=arm   go build -tags release -ldflags $(LD_FLAGS) -o _build/$(NAME)-$(VERSION)-linux-arm
	GOOS=linux   GOARCH=arm64 go build -tags release -ldflags $(LD_FLAGS) -o _build/$(NAME)-$(VERSION)-linux-arm64
	GOOS=windows GOARCH=amd64 go build -tags release -ldflags $(LD_FLAGS) -o _build/$(NAME)-$(VERSION)-windows-amd64.exe
	cd _build; sha256sum * > sha256sums.txt
	@echo "Build completed. Binaries are in _build/ directory."

# Build current platform binary to bin/ directory
build-local: clean
	@echo "Building for current platform "
	mkdir -p $(OUTPUT_DIR)
	go build -tags release -ldflags $(LD_FLAGS) -o $(OUTPUT_DIR)/$(NAME)
	@echo "Build completed. Binary is in $(OUTPUT_DIR)/ directory."

image:
	docker build -t $(NAME) -f Dockerfile .

# release: build-all
# 	@command -v gh >/dev/null 2>&1 || { \
# 		echo "Error: GitHub CLI (gh) is required but not installed."; \
# 		echo "Install it from: https://cli.github.com/"; \
# 		echo "Or run: brew install gh"; \
# 		exit 1; \
# 	}
# 	@echo "Releasing version $(VERSION)..."
# 	mkdir -p release
# 	cp _build/* release/
# 	cd release && sha256sum --quiet --check sha256sums.txt
# 	gh release create v$(VERSION) release/* --title "v$(VERSION)" --notes "Release v$(VERSION)" 
# 	@echo "Release v$(VERSION) created."


# 安装补全文件到 Homebrew 目录
install-completions: build
	@echo "Installing shell completions..."
	$(eval HOMEBREW_PREFIX := $(shell brew --prefix 2>/dev/null || echo "/usr/local"))
	@echo "Using prefix: $(HOMEBREW_PREFIX)"
	
	install -d $(HOMEBREW_PREFIX)/etc/bash_completion.d/
	install -m 644 completions/${NAME}.bash $(HOMEBREW_PREFIX)/etc/bash_completion.d/${NAME}
	
	install -d $(HOMEBREW_PREFIX)/share/zsh/site-functions/
	install -m 644 completions/${NAME}.zsh $(HOMEBREW_PREFIX)/share/zsh/site-functions/_${NAME}
	
	install -d $(HOMEBREW_PREFIX)/share/fish/vendor_completions.d/
	install -m 644 completions/${NAME}.fish $(HOMEBREW_PREFIX)/share/fish/vendor_completions.d/${NAME}.fish
	
	@echo "Completions installed successfully!"

# Install to local system (adjust path as needed)
install: build install-completions
	@echo "Installing $(NAME) to /usr/local/bin/..."
	sudo cp $(OUTPUT_DIR)/$(NAME) /usr/local/bin/$(NAME)
	@echo "Installation completed."

# Install to user's bin directory
install-user: build-local
	@echo "Installing $(NAME) to ~/bin/..."
	mkdir -p ~/bin
	cp $(OUTPUT_DIR)/$(NAME) ~/bin/$(NAME)
	@echo "Installation completed. Make sure ~/bin is in your PATH."


# Test GoReleaser configuration
goreleaser-check:
	@command -v goreleaser >/dev/null 2>&1 || { \
		echo "Error: goreleaser is required but not installed."; \
		echo "Install it from: https://goreleaser.com/install/"; \
		echo "Or run: brew install goreleaser"; \
		exit 1; \
	}
	goreleaser check

# Test build with GoReleaser (no release)
goreleaser-snapshot: clean goreleaser-check
	goreleaser release --snapshot

release: clean goreleaser-check
	@echo "Releasing version $(VERSION) with GoReleaser..."
	@export GITHUB_TOKEN=$$(gh auth token) && goreleaser release

# Show build information
# Show detailed build information
info:
	@echo "╔══════════════════════════════════════════════════════════════╗"
	@echo "║                    Build Information                         ║"
	@echo "╠══════════════════════════════════════════════════════════════╣"
	@echo "║  Project:    $(NAME)                                         ║"
	@echo "║  Version:    $(VERSION)                                      ║"
	@echo "║  Build Hash: $(BUILD)                                        ║"
	@echo "║  Output Dir: $(OUTPUT_DIR)                                   ║"
	@echo "╠══════════════════════════════════════════════════════════════╣"
	@echo "║  Go Version: $(shell go version | cut -d' ' -f3)             ║"
	@echo "║  Platform:   $(shell go env GOOS)/$(shell go env GOARCH)     ║"
	@echo "║  GOPATH:     $(shell go env GOPATH)                          ║"
	@echo "║  GOROOT:     $(shell go env GOROOT)                          ║"
	@echo "╠══════════════════════════════════════════════════════════════╣"
	@echo "║  Build Flags:                                                ║"
	@echo "║    - LD Flags: $(LD_FLAGS)                                   ║"
	@echo "║    - Build Tags: release                                     ║"
	@echo "╠══════════════════════════════════════════════════════════════╣"
	@echo "║  Available Targets:                                          ║"
	@echo "║    • build       - Build release binary                      ║"
	@echo "║    • build-dev   - Build development binary                  ║"
	@echo "║    • build-all   - Build for all platforms                   ║"
	@echo "║    • build-local - Build for current platform                ║"
	@echo "║    • release     - Create release with GoReleaser            ║"
	@echo "║    • install     - Install to system                         ║"
	@echo "║    • clean       - Clean build artifacts                     ║"
	@echo "╚══════════════════════════════════════════════════════════════╝"

# Alternative minimalist version
info-simple:
	@echo "📦 $(NAME) v$(VERSION) ($(BUILD))"
	@echo "📍 Output: $(OUTPUT_DIR)/"
	@echo "🛠️  Go: $(shell go version | cut -d' ' -f3)"
	@echo "💻 Platform: $(shell go env GOOS)/$(shell go env GOARCH)"

.PHONY: build build-dev build-all build-local clean image release install install-user info info-simple