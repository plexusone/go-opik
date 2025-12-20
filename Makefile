.PHONY: all test lint build docs presentation clean

all: test lint build

test:
	go test -v -race ./...

lint:
	golangci-lint run

build:
	go build ./...

# Build CLI for current platform
build-cli:
	go build -o bin/opik ./cmd/opik

# Generate Marp presentation HTML
presentation:
	npx @marp-team/marp-cli --theme vibeminds.css --html presentation.md -o docsrc/presentation.html

# Build MkDocs site (requires mkdocs installed)
docs: presentation
	mkdocs build

# Serve MkDocs locally for development
docs-serve: presentation
	mkdocs serve

clean:
	rm -rf bin/ site/ docsrc/presentation.html

help:
	@echo "Available targets:"
	@echo "  test         - Run tests with race detection"
	@echo "  lint         - Run golangci-lint"
	@echo "  build        - Build all packages"
	@echo "  build-cli    - Build CLI binary"
	@echo "  presentation - Generate Marp presentation HTML"
	@echo "  docs         - Build MkDocs site (includes presentation)"
	@echo "  docs-serve   - Serve MkDocs locally"
	@echo "  clean        - Remove build artifacts"
