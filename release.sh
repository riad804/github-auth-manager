#!/bin/bash
VERSION="v1.0.0"
mkdir -p release/$VERSION

platforms=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name="mycli-$VERSION-$GOOS-$GOARCH"
    
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    env GOOS=$GOOS GOARCH=$GOARCH go build -o release/$VERSION/$output_name
    shasum -a 256 release/$VERSION/$output_name > release/$VERSION/$output_name.sha256
done