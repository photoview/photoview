#!/bin/bash
set -euo pipefail

# Fallback to the latest version if JELLYFIN_FFMPEG_VERSION is not set
if [[ -z "$JELLYFIN_FFMPEG_VERSION" ]]; then
  echo "WARN: jellyfin-ffmpeg version is empty, most likely the script runs not on CI."
  echo "Fetching the latest version from jellyfin-ffmpeg repo..."
  JELLYFIN_FFMPEG_VERSION=$(curl -fsSL --retry 2 --retry-delay 5 --retry-max-time 60 \
    "https://api.github.com/repos/jellyfin/jellyfin-ffmpeg/releases/latest" | jq -r '.tag_name')
fi

: "${DEB_HOST_MULTIARCH:=$(uname -m)-linux-gnu}"
: "${DEB_HOST_ARCH:=$(dpkg --print-architecture)}"
CACHE_DIR="${BUILD_CACHE_DIR:-/build-cache}/jellyfin-ffmpeg-${JELLYFIN_FFMPEG_VERSION}"
CACHE_MARKER="${CACHE_DIR}/jellyfin-ffmpeg-${JELLYFIN_FFMPEG_VERSION}-complete"

# Check if this specific version is already downloaded and cached
if [[ -f "$CACHE_MARKER" ]] && [[ -d "${CACHE_DIR}/output" ]]; then
  echo "jellyfin-ffmpeg ${JELLYFIN_FFMPEG_VERSION} found in cache, reusing..."
  mkdir -p /output
  cp -ra "${CACHE_DIR}/output/"* /output/
  exit 0
fi

echo "Downloading jellyfin-ffmpeg ${JELLYFIN_FFMPEG_VERSION} (cache miss)..."

echo Compiler: "${DEB_HOST_MULTIARCH}" Arch: "${DEB_HOST_ARCH}"

VER="${JELLYFIN_FFMPEG_VERSION#v}"
MAJOR_VER=$(echo "${VER}" | cut -d. -f1)
URL="https://github.com/jellyfin/jellyfin-ffmpeg/releases/download/${JELLYFIN_FFMPEG_VERSION}/jellyfin-ffmpeg${MAJOR_VER}_${VER}-bookworm_${DEB_HOST_ARCH}.deb"
echo download jellyfin-ffmpeg from "$URL"
mkdir -p /output/deb
curl -fsSL --retry 2 --retry-delay 5 --retry-max-time 60 -o /output/deb/jellyfin-ffmpeg.deb "$URL"

# After successful download, cache the results
echo "Caching jellyfin-ffmpeg ${JELLYFIN_FFMPEG_VERSION} build results..."
mkdir -p "${CACHE_DIR}/output"
cp -ra /output/* "${CACHE_DIR}/output/"
touch "$CACHE_MARKER"

echo "jellyfin-ffmpeg ${JELLYFIN_FFMPEG_VERSION} downloaded and cached"
