#!/bin/bash

# Define the list of target platforms
PLATFORMS=("darwin-amd64" "linux-386" "linux-amd64" "windows-386" "windows-amd64")

# Create the dist directory if it doesnt exist
mkdir -p dist

for PLATFORM in "${PLATFORMS[@]}"; do
    # Extract the GOOS and GOARCH values from the platform string
    GOOS=${PLATFORM%%-*}
    GOARCH=${PLATFORM##*-}
    echo "Building for $GOOS/$GOARCH..."

    # Set the GOOS and GOARCH environment variables
    export GOOS
    export GOARCH

    # Determine the binary extension based on the target OS
    BIN_EXT=""
    if [[ "$GOOS" == "windows" ]]; then
        BIN_EXT=".exe"
    fi

    # Build the executable
    go build -o "dist/JSONUIHELPER-$GOOS-$GOARCH$BIN_EXT"
done

echo "Build completed."