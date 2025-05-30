#!/bin/bash

set -e

APP_NAME="gham"  # Change this to your CLI tool name
VERSION="$1"

if [ -z "$VERSION" ]; then
  echo "❌ Error: Version not provided."
  echo "Usage: ./release.sh v1.0.0"
  exit 1
fi

DIST_DIR="dist/$VERSION"

# List of target OS/ARCH combinations
targets=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
  "windows/arm64"
)

# Clean old build
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"

echo "🚀 Building $APP_NAME version $VERSION for multiple platforms..."

for target in "${targets[@]}"; do
  IFS="/" read -r GOOS GOARCH <<< "$target"
  output_name="${APP_NAME}-${VERSION}-${GOOS}-${GOARCH}"
  [ "$GOOS" = "windows" ] && output_name+=".exe"

  echo "-> Building for $GOOS/$GOARCH..."

  env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$VERSION" -o "$DIST_DIR/$output_name" .

done

echo "✅ Builds completed in '$DIST_DIR'."
