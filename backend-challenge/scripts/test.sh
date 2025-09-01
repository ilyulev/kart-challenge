#!/bin/bash
set -e

echo "Running Oolio Food API tests..."

# Run unit tests
echo "Running unit tests..."
go test -v ./...

# Run tests with race detection
echo "Running tests with race detection..."
go test -race -v ./...

# Generate coverage report
echo "Generating coverage report..."
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

echo "Tests completed successfully!"
echo "Coverage report: coverage.html"