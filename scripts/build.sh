#!/bin/bash

# Build the IM System server
echo "Building IM System server..."

# Create bin directory if it doesn't exist
mkdir -p bin

# Build the server
go build -o bin/im-server ./cmd

echo "Build completed. Binary saved to bin/im-server"
