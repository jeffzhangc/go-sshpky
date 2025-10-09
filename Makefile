NAME=sshpky
VERSION=$(shell cat VERSION)
BUILD=$(shell git rev-parse --short HEAD)
EXT_LD_FLAGS="-Wl"
LD_FLAGS="-s -w -X main.version=$(VERSION) -X main.build=$(BUILD) -extldflags=$(EXT_LD_FLAGS)"
OUTPUT_DIR=./bin

clean:
	rm -rf _build/ release/ $(OUTPUT_DIR)/

build:
	@echo "Building "
	go mod tidy
	go build -tags release -ldflags $(LD_FLAGS) -o $(OUTPUT_DIR)/$(NAME)

build-dev:
	@echo "Building development version "
	mkdir -p $(OUTPUT_DIR)
	go build -ldflags "-w -X main.version=$(VERSION)-dev -X main.build=$(BUILD) -extldflags=$(EXT_LD_FLAGS)" -o $(OUTPUT_DIR)/$(NAME)-dev

build-all: clean
	@echo "Building for all platforms from $(SRC_DIR)..."
	mkdir -p _build $(OUTPUT_DIR)
# 	GOOS=darwin  GOARCH=arm64 go build -tags release -ldflags $(LD_FLAGS) -o _build/$(NAME)-$(VERSION)-darwin-arm64
	GOOS=darwin  GOARCH=amd64 go build -tags release -ldflags $(LD_FLAGS) -o _build/$(NAME)-$(VERSION)-darwin-amd64
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

release: build-all
	@command -v gh >/dev/null 2>&1 || { \
		echo "Error: GitHub CLI (gh) is required but not installed."; \
		echo "Install it from: https://cli.github.com/"; \
		echo "Or run: brew install gh"; \
		exit 1; \
	}
	@echo "Releasing version $(VERSION)..."
	mkdir -p release
	cp _build/* release/
	cd release && sha256sum --quiet --check sha256sums.txt
	gh release create v$(VERSION) release/* --title "v$(VERSION)" --notes "Release v$(VERSION)" 
	@echo "Release v$(VERSION) created."

# Install to local system (adjust path as needed)
install: build-local
	@echo "Installing $(NAME) to /usr/local/bin/..."
	sudo cp $(OUTPUT_DIR)/$(NAME) /usr/local/bin/$(NAME)
	@echo "Installation completed."

# Install to user's bin directory
install-user: build-local
	@echo "Installing $(NAME) to ~/bin/..."
	mkdir -p ~/bin
	cp $(OUTPUT_DIR)/$(NAME) ~/bin/$(NAME)
	@echo "Installation completed. Make sure ~/bin is in your PATH."

# Show build information
info:
	@echo "Project: $(NAME)"
	@echo "Version: $(VERSION)"
	@echo "Build: $(BUILD)"
	@echo "Output Directory: $(OUTPUT_DIR)"

.PHONY: build build-dev build-all build-local clean image release install install-user info