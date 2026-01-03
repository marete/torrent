#!/bin/bash
#
# Run static analysis tools on the codebase.
# Exit codes: 0 = success, non-zero = failure
#

set -e

echo "=== Verifying module dependencies ==="
go mod verify
go mod download
echo "Dependencies verified."

echo ""
echo "=== Running go vet ==="
echo "Go version: $(go version)"
go vet ./...
echo "go vet passed (no issues found)."

echo ""
echo "=== Running staticcheck ==="
if ! command -v staticcheck &> /dev/null; then
    echo "Installing staticcheck..."
    go install honnef.co/go/tools/cmd/staticcheck@latest
fi
echo "staticcheck version: $(staticcheck -version)"
staticcheck ./...
echo "staticcheck passed (no issues found)."

echo ""
echo "=== Running golangci-lint ==="
if ! command -v golangci-lint &> /dev/null; then
    echo "Installing golangci-lint v2..."
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
fi
echo "golangci-lint version: $(golangci-lint version --short)"
golangci-lint run ./...
echo "golangci-lint passed."

echo ""
echo "=== Checking code formatting ==="
echo "Checking with gofmt..."
UNFORMATTED=$(gofmt -l .)
if [ -n "$UNFORMATTED" ]; then
    echo "The following files are not formatted correctly:"
    echo "$UNFORMATTED"
    exit 1
fi
echo "gofmt passed (all files formatted correctly)."

echo ""
echo "=== All static analysis checks passed ==="
