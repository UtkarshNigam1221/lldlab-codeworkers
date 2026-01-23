#!/bin/bash

set -e

echo "========================================="
echo "Running Code Quality Checks"
echo "========================================="

echo ""
echo "1. Running go fmt..."
if ! go fmt ./...; then
    echo "❌ go fmt failed"
    exit 1
fi
echo "✅ go fmt passed"

echo ""
echo "2. Running go vet..."
if ! go vet ./...; then
    echo "❌ go vet failed"
    exit 1
fi
echo "✅ go vet passed"

echo ""
echo "3. Running unit tests..."
if ! go test -v -short -race -coverprofile=coverage.out ./...; then
    echo "❌ Unit tests failed"
    exit 1
fi
echo "✅ Unit tests passed"

echo ""
echo "4. Generating coverage report..."
go tool cover -func=coverage.out | tail -n 1

echo ""
echo "========================================="
echo "✅ All checks passed!"
echo "========================================="

# Display coverage summary
echo ""
echo "Coverage Summary:"
go tool cover -func=coverage.out | grep total
