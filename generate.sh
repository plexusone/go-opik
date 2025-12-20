#!/bin/bash
# generate.sh - Generate Go API client from OpenAPI specification using ogen
#
# Prerequisites:
#   go install github.com/ogen-go/ogen/cmd/ogen@latest
#
# Usage:
#   ./generate.sh
#
# This script:
#   1. Runs ogen to generate Go code from openapi/openapi.yaml
#   2. Applies fixes for known issues (jx.Raw comparison)
#   3. Runs go mod tidy to update dependencies
#   4. Verifies the build compiles
#
# To update the OpenAPI spec before running:
#   curl -o openapi/openapi.yaml \
#     https://raw.githubusercontent.com/comet-ml/opik/main/sdks/code_generation/fern/openapi/openapi.yaml

set -e

# Check if ogen is installed
if ! command -v ogen &> /dev/null; then
    echo "Error: ogen is not installed."
    echo "Install with: go install github.com/ogen-go/ogen/cmd/ogen@latest"
    exit 1
fi

# Check if OpenAPI spec exists
if [ ! -f "openapi/openapi.yaml" ]; then
    echo "Error: openapi/openapi.yaml not found."
    echo "Download it with:"
    echo "  curl -o openapi/openapi.yaml \\"
    echo "    https://raw.githubusercontent.com/comet-ml/opik/main/sdks/code_generation/fern/openapi/openapi.yaml"
    exit 1
fi

echo "Generating API code with ogen..."
ogen --package api --target internal/api --clean openapi/openapi.yaml

echo "Fixing jx.Raw comparison issues in equal files..."
# Fix: ogen generates code that compares jx.Raw values directly, which doesn't
# work in Go. We need to use bytes.Equal() instead.
if [ -f internal/api/oas_experiment_item_equal_gen.go ]; then
    # Add bytes import
    sed -i '' 's/import "github.com\/ogen-go\/ogen\/validate"/import (\n\t"bytes"\n\n\t"github.com\/ogen-go\/ogen\/validate"\n)/' internal/api/oas_experiment_item_equal_gen.go
    # Replace direct comparisons with bytes.Equal
    sed -i '' 's/if a.Input != b.Input {/if !bytes.Equal([]byte(a.Input), []byte(b.Input)) {/' internal/api/oas_experiment_item_equal_gen.go
    sed -i '' 's/if a.Output != b.Output {/if !bytes.Equal([]byte(a.Output), []byte(b.Output)) {/' internal/api/oas_experiment_item_equal_gen.go
    echo "  Fixed oas_experiment_item_equal_gen.go"
fi

echo "Running go mod tidy..."
go mod tidy

echo "Verifying build..."
go build ./...

echo ""
echo "Done! API client regenerated successfully."
echo ""
echo "Next steps:"
echo "  1. Review changes in internal/api/"
echo "  2. Update SDK wrapper code if needed for new/changed endpoints"
echo "  3. Run tests: go test ./..."
echo "  4. Run linter: golangci-lint run"
