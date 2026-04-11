#!/bin/bash

set -euo pipefail

echo ${BASH_SOURCE[0]}

echo "Fetching the latest version from LibRaw repo..."
LIBRAW_VERSION=$(curl -fsSL --retry 2 --retry-delay 5 --retry-max-time 60 \
  "https://api.github.com/repos/LibRaw/LibRaw/releases/latest" | grep tag_name | sed 's/.*: "\(.*\)",*/\1/')
echo libraw version: ${LIBRAW_VERSION}

echo "Fetching the latest version from ImageMagick repo..."
IMAGEMAGICK_VERSION=$(curl -fsSL --retry 2 --retry-delay 5 --retry-max-time 60 \
  "https://api.github.com/repos/ImageMagick/ImageMagick/releases/latest" | grep tag_name | sed 's/.*: "\(.*\)",*/\1/')
echo imagemagick version: ${IMAGEMAGICK_VERSION}

echo "Fetching the latest version from jellyfin-ffmpeg repo..."
JELLYFIN_FFMPEG_VERSION=$(curl -fsSL --retry 2 --retry-delay 5 --retry-max-time 60 \
  "https://api.github.com/repos/jellyfin/jellyfin-ffmpeg/releases/latest" | grep tag_name | sed 's/.*: "\(.*\)",*/\1/')
echo jellyfin-ffmpeg version: ${JELLYFIN_FFMPEG_VERSION}

BUILD_ARGS=""
if [[ $# = "1" ]] && [[ -n "$1" ]]; then
  BUILD_ARGS="-t $1"
fi

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"

docker build \
  --build-arg "LIBRAW_VERSION=${LIBRAW_VERSION}" \
  --build-arg "IMAGEMAGICK_VERSION=${IMAGEMAGICK_VERSION}" \
  --build-arg "JELLYFIN_FFMPEG_VERSION=${JELLYFIN_FFMPEG_VERSION}" \
  ${BUILD_ARGS} \
  "${SCRIPT_DIR}"
