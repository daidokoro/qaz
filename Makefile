#!/bin/bash

set -e

# Build the CLI
echo "Building qaz..."
go build -o qaz -ldflags "-s -w" ./cmd/root

echo "qaz binary built successfully in the root directory."