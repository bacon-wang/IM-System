#!/bin/bash

# Build the IM System
echo "Building IM System..."

# Create bin directory if it doesn't exist
mkdir -p bin

# Build the server
echo "Building server..."
go build -o bin/im-server ./cmd

# Build the client tool
echo "Building client tool..."
go build -o bin/im-client ./tools/client

echo "Build completed!"
echo "  Server binary: bin/im-server"
echo "  Client tool:   bin/im-client"
