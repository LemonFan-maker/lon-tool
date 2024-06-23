# Define variables
NAME = lon-tool
SRC ?= main.go
BIN_DIR ?= ./bin
GIT_VERSION := $(shell git describe --abbrev=4 --dirty --always --tags)
GOFLAGS ?=  -ldflags="-s -w -X lon-tool/cmd.version=$(GIT_VERSION)"
WINDOWS_GOFLAGS ?=  -ldflags="-extldflags=-static -s -w -X lon-tool/cmd.version=$(GIT_VERSION)"

# Windows settings
WINDOWS_BIN = $(BIN_DIR)/$(NAME)_win_amd64.exe
WINDOWS_CC = x86_64-w64-mingw32-gcc
WINDOWS_CXX = x86_64-w64-mingw32-g++
WINDOWS_PKG_CONFIG_PATH = /usr/x86_64-w64-mingw32/lib/pkgconfig
WINDOWS_CGO_CFLAGS = -I/usr/x86_64-w64-mingw32/include
WINDOWS_CGO_LDFLAGS = -L/usr/x86_64-w64-mingw32/lib

# Linux settings
LINUX_BIN = $(BIN_DIR)/$(NAME)_lin_amd64
LINUX_CC = x86_64-linux-gnu-gcc
LINUX_CXX = x86_64-linux-gnu-g++
LINUX_PKG_CONFIG_PATH = /usr/lib/pkgconfig
LINUX_CGO_CFLAGS = -I/usr/include
LINUX_CGO_LDFLAGS = -L/usr/lib

# Targets
all: windows linux

windows:
	@echo "Building windows bin"
	@export CGO_ENABLED=1 GOARCH=amd64 GOOS=windows CC=$(WINDOWS_CC) CXX=$(WINDOWS_CXX) \
	PKG_CONFIG_PATH=$(WINDOWS_PKG_CONFIG_PATH) CGO_CFLAGS=$(WINDOWS_CGO_CFLAGS) \
	CGO_LDFLAGS=$(WINDOWS_CGO_LDFLAGS) && \
	go build $(WINDOWS_GOFLAGS) -o $(WINDOWS_BIN) $(SRC) && \
	echo "- saved to $(WINDOWS_BIN)"

linux:
	@echo "Building linux bin"
	@export CGO_ENABLED=1 GOARCH=amd64 GOOS=linux CC=$(LINUX_CC) CXX=$(LINUX_CXX) \
	PKG_CONFIG_PATH=$(LINUX_PKG_CONFIG_PATH) CGO_CFLAGS=$(LINUX_CGO_CFLAGS) \
	CGO_LDFLAGS=$(LINUX_CGO_LDFLAGS) && \
	go build $(GOFLAGS) -o $(LINUX_BIN) $(SRC) && \
	echo "- saved to $(LINUX_BIN)"

clean:
	rm -f $(MACOS_BIN) $(WINDOWS_BIN) $(LINUX_BIN)

.PHONY: all macos windows linux clean
