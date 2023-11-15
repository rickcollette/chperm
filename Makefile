.PHONY: all build directories config deb rpm shar clean

# Set the output binary name
BINARY_NAME=chperm
VERSION ?= 1.0.0 # Default version
BUILD_ID=$(shell git rev-parse --short HEAD)

all: directories build config

directories:
	@echo "Creating directory structure..."
	@mkdir -p dist/usr/local/bin
	@mkdir -p dist/usr/local/etc
	@mkdir -p dist/usr/local/share/rollback
	@mkdir -p dist/usr/share/man/man1
	@chmod 755 buildpkgs.sh

build:
	@echo "Building the Go binary..."
	@go build -ldflags "-buildid=$(BUILD_ID) -o dist/usr/local/bin/$(BINARY_NAME)
	
config:
	@echo "Copying configuration file..."
	@cp chperm.conf dist/usr/local/etc/chperm.conf
	@cp chperm.1 dist/usr/share/man/man1/chperm.1

clean:
	@echo "Cleaning up..."
	@rm -rf dist build

deb: all
	@echo "Building Debian package..."
	@echo "Version: $(VERSION)"
	@exec ./buildpkgs.sh create_deb $(VERSION)

rpm: all
	@echo "Building RPM package..."
	@exec ./buildpkgs.sh create_rpm $(VERSION)

shar: all
	@echo "Building SHAR package..."
	@exec ./buildpkgs.sh create_shar $(VERSION)

packages: deb rpm shar
	@echo "All packages built!"
