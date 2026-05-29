#!/bin/bash

# Exit on error
set -e

echo "Building anvil..."
go build -o anvil main.go

echo "Installing to /usr/local/bin..."
sudo mv anvil /usr/local/bin/anvil

echo "Applying macOS code signature..."
sudo codesign --force --deep --sign - /usr/local/bin/anvil

echo "Successfully installed anvil!"
anvil --help
