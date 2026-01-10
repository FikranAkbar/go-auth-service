#!/bin/bash

# Air runner script
# This script ensures Air is available and runs it

AIR_BIN="air"

# Check if air is in PATH
if ! command -v air &> /dev/null; then
    # Try using GOPATH
    if [ -f "$HOME/go/bin/air" ]; then
        AIR_BIN="$HOME/go/bin/air"
    elif [ -f "$(go env GOPATH)/bin/air" ]; then
        AIR_BIN="$(go env GOPATH)/bin/air"
    else
        echo "Air is not installed. Installing..."
        go install github.com/air-verse/air@latest
        AIR_BIN="$(go env GOPATH)/bin/air"
    fi
fi

echo "Starting application with Air (hot reload)..."
$AIR_BIN

