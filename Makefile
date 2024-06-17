# Define variables
GOFLAGS ?=  -ldflags="-s -w"
SRC ?= main.go
BIN_DIR ?= ./bin

# macOS settings
MACOS_BIN = $(BIN_DIR)/lnd_mac_amd64
MACOS_CC = x86_64-apple-darwin23-gcc-14
MACOS_CXX = x86_64-apple-darwin23-g++-14

# Windows settings
WINDOWS_BIN = $(BIN_DIR)/lnd_win_amd64.exe
WINDOWS_CC = x86_64-w64-mingw32-gcc
WINDOWS_CXX = x86_64-w64-mingw32-g++
WINDOWS_PKG_CONFIG_PATH = /usr/local/x86_64-w64-mingw32/lib/pkgconfig
WINDOWS_CGO_CFLAGS = -I/usr/local/x86_64-w64-mingw32/include
WINDOWS_CGO_LDFLAGS = -L/usr/local/x86_64-w64-mingw32/lib

# Linux settings
LINUX_BIN = $(BIN_DIR)/lnd_lin_amd64
LINUX_CC = x86_64-linux-gnu-gcc
LINUX_CXX = x86_64-linux-gnu-g++
LINUX_PKG_CONFIG_PATH = /usr/local/x86_64-linux-gnu/lib/pkgconfig
LINUX_CGO_CFLAGS = -I/usr/local/x86_64-linux-gnu/include
LINUX_CGO_LDFLAGS = -L/usr/local/x86_64-linux-gnu/lib

# Targets
all: macos windows linux

macos:
	@export CGO_ENABLED=1 GOARCH=amd64 GOOS=darwin CC=$(MACOS_CC) CXX=$(MACOS_CXX) && \
	echo "Building macos bin" && \
	go build $(GOFLAGS) -o $(MACOS_BIN) $(SRC) && \
	echo "- saved to $(MACOS_BIN)"

windows:
	@export CGO_ENABLED=1 GOARCH=amd64 GOOS=windows CC=$(WINDOWS_CC) CXX=$(WINDOWS_CXX) \
	PKG_CONFIG_PATH=$(WINDOWS_PKG_CONFIG_PATH) CGO_CFLAGS=$(WINDOWS_CGO_CFLAGS) \
	CGO_LDFLAGS=$(WINDOWS_CGO_LDFLAGS) && \
	echo "Building windows bin" && \
	go build $(GOFLAGS) -o $(WINDOWS_BIN) -ldflags="-extldflags=-static" $(SRC) && \
	echo "- saved to $(WINDOWS_BIN)"

linux:
	@export CGO_ENABLED=1 GOARCH=amd64 GOOS=linux CC=$(LINUX_CC) CXX=$(LINUX_CXX) \
	PKG_CONFIG_PATH=$(LINUX_PKG_CONFIG_PATH) CGO_CFLAGS=$(LINUX_CGO_CFLAGS) \
	CGO_LDFLAGS=$(LINUX_CGO_LDFLAGS) && \
	echo "Building linux bin" && \
	go build $(GOFLAGS) -o $(LINUX_BIN) $(SRC) && \
	echo "- saved to $(LINUX_BIN)"

clean:
	rm -f $(MACOS_BIN) $(WINDOWS_BIN) $(LINUX_BIN)

.PHONY: all macos windows linux clean
