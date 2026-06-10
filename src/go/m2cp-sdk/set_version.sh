#!/bin/bash

# set_version.sh

# Navigate to the directory of the script
cd "$(dirname "$0")"

# Check if a new version number is provided
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 new_version"
    exit 1
fi

new_version=$1

# Use sed to replace the version number in version.go
sed -i "s/^const version = \".*\"$/const version = \"$new_version\"/" version.go

# Verify and echo the change
echo "Updated version to $new_version in version.go"
