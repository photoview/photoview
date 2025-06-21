#!/bin/bash
set -euo pipefail

: "${DEB_HOST_MULTIARCH:=x86_64-linux-gnu}"
: "${DEB_HOST_ARCH:=amd64}"
CACHE_DIR="${BUILD_CACHE_DIR:-/build-cache}/jellyfin-ffmpeg-${JELLYFIN_FFMPEG_VERSION}"
CACHE_MARKER="${CACHE_DIR}/jellyfin-ffmpeg-${JELLYFIN_FFMPEG_VERSION}-complete"

# Check if this specific version is already downloaded and cached
if [[ -f "$CACHE_MARKER" ]] && [[ -d "${CACHE_DIR}/output" ]]; then
  echo "jellyfin-ffmpeg ${JELLYFIN_FFMPEG_VERSION} found in cache, reusing..."
  mkdir -p /output
  cp -r "${CACHE_DIR}/output/"* /output/
  exit 0
fi

echo "Downloading jellyfin-ffmpeg ${JELLYFIN_FFMPEG_VERSION} (cache miss)..."

echo Compiler: "${DEB_HOST_MULTIARCH}" Arch: "${DEB_HOST_ARCH}"

URL="https://api.github.com/repos/jellyfin/jellyfin-ffmpeg/tarball/${JELLYFIN_FFMPEG_VERSION}"
echo download jellyfin-ffmpeg from "$URL"
mkdir -p /output/deb
curl -L -o /output/deb/jellyfin-ffmpeg.deb ${GITHUB_TOKEN:+-H "Authorization: Bearer ${GITHUB_TOKEN}"} "$URL"

# After successful download, cache the results
echo "Caching jellyfin-ffmpeg ${JELLYFIN_FFMPEG_VERSION} build results..."
mkdir -p "${CACHE_DIR}/output"
cp -r /output/* "${CACHE_DIR}/output/"
touch "$CACHE_MARKER"

echo "jellyfin-ffmpeg ${JELLYFIN_FFMPEG_VERSION} downloaded and cached"
