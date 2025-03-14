#!/bin/bash
set -e  # Stops script on first error
PWD=$(pwd)  # Get current working directory
Dir="../kwil"  # Relative path

# Get absolute path
AbsPath=$(realpath "$PWD/$Dir")

# Ensure the directory exists before attempting to delete
if [[ -d "$AbsPath" ]]; then
    echo "Removing directory: $AbsPath"
    rm -rf "$AbsPath"
else
    echo "Directory does not exist: $AbsPath"
fi
