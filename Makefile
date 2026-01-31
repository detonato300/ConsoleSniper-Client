# Makefile for ConsoleSniper Client

BINARY_NAME=consolesniper
CMD_PATH=./cmd/client/main.go
BUILD_FLAGS=-ldflags="-s -w"

# Garble settings for production builds
# Note: -literals may break some experimental JSON libraries.
# If build fails, try removing -literals or switching to standard encoding/json.
GARBLE_FLAGS=-tiny

.PHONY: build clean test check prod-build cross-compile

build:
	go build $(BUILD_FLAGS) -o $(BINARY_NAME).exe $(CMD_PATH)

prod-build:
	@echo "Building hardened production binary (Windows)..."
	garble $(GARBLE_FLAGS) build $(BUILD_FLAGS) -o $(BINARY_NAME)_prod.exe $(CMD_PATH)

cross-compile:
	@echo "Building production binaries for multiple platforms..."
	# Windows x64
	GOOS=windows GOARCH=amd64 garble $(GARBLE_FLAGS) build $(BUILD_FLAGS) -o $(BINARY_NAME)_win_x64.exe $(CMD_PATH)
	# Linux x64
	GOOS=linux GOARCH=amd64 garble $(GARBLE_FLAGS) build $(BUILD_FLAGS) -o $(BINARY_NAME)_linux_x64 $(CMD_PATH)

clean:
	if exist *.exe del *.exe
	if exist consolesniper_linux_x64 del consolesniper_linux_x64

test:
	go test -v ./...

check: test build
	@echo "Build and tests successful."