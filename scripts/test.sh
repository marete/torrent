#!/bin/bash
#
# Run all tests with various instrumentation modes.
# Exit codes: 0 = success, non-zero = failure
#

set -e

echo "=== Building project ==="
go build -v ./...

echo ""
echo "=== Running unit tests ==="
go test -v -timeout 5m ./...

echo ""
echo "=== Running tests with race detector ==="
go test -race -timeout 10m ./...

echo ""
echo "=== Running tests with coverage ==="
go test -coverprofile=coverage.out -covermode=atomic -timeout 5m ./...
echo ""
echo "Coverage summary:"
go tool cover -func=coverage.out | tail -1

echo ""
echo "=== All tests passed ==="
