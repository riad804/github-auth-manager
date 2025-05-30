#!/bin/bash
set -euo pipefail

# Get version from first argument or use git tag
VERSION=${1:-$(git describe --tags --abbrev=0)}
if [ -z "$VERSION" ]; then
  echo "Error: No version specified and no git tag found"
  exit 1
fi

echo "Building GHAM version $VERSION"
mkdir -p release/$VERSION

platforms=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
)

for platform in "${platforms[@]}"; do
  platform_split=(${platform//\// })
  GOOS=${platform_split[0]}
  GOARCH=${platform_split[1]}
  output_name="gham-$VERSION-$GOOS-$GOARCH"
  
  if [ "$GOOS" = "windows" ]; then
    output_name+='.exe'
  fi

  echo "Building $output_name"
  CGO_ENABLED=1 GOOS=$GOOS GOARCH=$GOARCH go build \
    -ldflags="-X main.AppVersion=$VERSION" \
    -o "release/$VERSION/$output_name" .
  
  # Generate checksums
  sha256sum "release/$VERSION/$output_name" > "release/$VERSION/$output_name.sha256"
done

echo "Build complete. Files saved to release/$VERSION/"