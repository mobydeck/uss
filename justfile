#!/usr/bin/env just --justfile

# Get version from most recent git tag, fallback to 'dev'
_get-version := `git describe --tags --abbrev=0 2>/dev/null || echo 'dev'`

# Default recipe
default:
    @just --list

# Run tests
test:
    go test -v ./...

# Run tests with coverage
test-coverage:
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

# Build binary for current platform
build:
    go build -trimpath -ldflags="-s -w -X github.com/mobydeck/uss/cmd/uss.version={{ _get-version }}" -o uss ./cmd/uss
    @echo "Built: uss (version: {{ _get-version }})"

# Helper function to build a binary for a specific OS/ARCH
_build-binary os arch:
    mkdir -p dist
    GOOS={{ os }} GOARCH={{ arch }} go build -trimpath -ldflags="-s -w -X github.com/mobydeck/uss/cmd/uss.version={{ _get-version }}" -o dist/uss-{{ os }}-{{ arch }} ./cmd/uss
    @echo "Built: dist/uss-{{ os }}-{{ arch }} (version: {{ _get-version }})"

# Build for Linux AMD64
build-linux-amd64: (_build-binary "linux" "amd64")

# Build for Linux ARM64
build-linux-arm64: (_build-binary "linux" "arm64")

# Build for macOS AMD64
build-darwin-amd64: (_build-binary "darwin" "amd64")

# Build for macOS ARM64
build-darwin-arm64: (_build-binary "darwin" "arm64")

# Build all distributions
build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64
    @echo "All distributions built successfully"

# Create distribution archives
dist: build-all
    #!/bin/bash
    set -e
    mkdir -p dist/archives
    cd dist

    # Create tarballs
    tar -czf archives/uss-linux-amd64.tar.gz uss-linux-amd64 ../README.md ../LICENSE
    tar -czf archives/uss-linux-arm64.tar.gz uss-linux-arm64 ../README.md ../LICENSE
    tar -czf archives/uss-darwin-amd64.tar.gz uss-darwin-amd64 ../README.md ../LICENSE
    tar -czf archives/uss-darwin-arm64.tar.gz uss-darwin-arm64 ../README.md ../LICENSE

    # Create checksums
    #cd archives && shasum -a 256 * > SHA256SUMS && cd ..

    echo "Distribution files created in dist/archives/"
    ls -lh archives/

# Create GitHub release (auto-detect version from git tag, or provide version)
# Usage: just release  (uses latest tag)  or  just release v0.1.0 (creates new tag)
release version='':
    #!/bin/bash
    set -e

    if [ -z "{{ version }}" ]; then
        VERSION="{{ _get-version }}"
        echo "Using auto-detected version: ${VERSION}"
    else
        VERSION="{{ version }}"
        # Check if tag exists
        if git rev-parse "${VERSION}" >/dev/null 2>&1; then
            echo "Tag ${VERSION} already exists"
            exit 1
        fi

        # Create the tag
        git tag -a "${VERSION}" -m "Release ${VERSION}"
        git push origin "${VERSION}"
        echo "Created git tag: ${VERSION}"
    fi

    # Create GitHub release
    gh release create "${VERSION}" \
        --title "Release ${VERSION}" \
        --notes "See README.md for installation and usage instructions."

    echo "GitHub release created: ${VERSION}"

# Upload dist files to GitHub release
# Usage: just upload-release  (uses latest tag)  or  just upload-release v0.1.0
upload-release version='':
    #!/bin/bash
    set -e

    if [ -z "{{ version }}" ]; then
        VERSION="{{ _get-version }}"
    else
        VERSION="{{ version }}"
    fi

    if [ ! -d "dist/archives" ]; then
        echo "Error: dist/archives not found. Run 'just dist' first"
        exit 1
    fi

    echo "Uploading files to release ${VERSION}..."
    cd dist/archives
    FILES=$(ls uss-*-*.tar.gz 2>/dev/null | head -4)

    if [ -z "$FILES" ]; then
        echo "Error: No distribution files found in dist/archives"
        exit 1
    fi

    gh release upload "${VERSION}" $FILES --clobber

    echo "Files uploaded to release ${VERSION}"
    echo "View release: $(gh release view ${VERSION} --json url -q .url)"

# Full release workflow: test, build, dist, create release, upload
# Usage: just full-release  (uses latest tag)  or  just full-release v0.1.0 (creates new release)
full-release version='':
    #!/bin/bash
    set -e

    if [ -z "{{ version }}" ]; then
        VERSION="{{ _get-version }}"
        echo "🔄 Full release workflow using version: ${VERSION}"
    else
        VERSION="{{ version }}"
        echo "🔄 Full release workflow creating new release: ${VERSION}"
    fi

    just test
    just build-all
    just dist
    just release {{ version }}
    just upload-release {{ version }}

    echo ""
    echo "✅ Release ${VERSION} complete!"
    echo "   GitHub: $(gh release view ${VERSION} --json url -q .url)"

# Clean build artifacts
clean:
    rm -f uss
    rm -rf dist/
    rm -f coverage.out coverage.html
    @echo "Cleaned build artifacts"

# Format code
fmt:
    go fmt ./...
    @echo "Code formatted"

# Run linter
lint:
    go vet ./...
    @echo "Lint check passed"

# Generate test coverage report
coverage:
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out
    @echo "Coverage report generated and opened"

# Development workflow: test, build, run
dev: test build
    ./uss --help

# Check dependencies
deps:
    go mod download
    go mod verify
    @echo "Dependencies verified"

# Update dependencies
update-deps:
    go get -u ./...
    go mod tidy
    @echo "Dependencies updated"

# Install binary to $GOPATH/bin
install: build
    cp uss ~/go/bin/uss
    @echo "Installed uss to ~/go/bin/uss"

# Show help
help:
    @just --list --unsorted
