#!/bin/bash

# Fallback to the latest version if LIBRAW_VERSION is not set
if [[ -z "${LIBRAW_VERSION}" ]]; then
  echo "WARN: LibRaw version is empty, most likely the script runs not on CI."
  echo "Fetching the latest version from LibRaw repo..."
  LIBRAW_VERSION=$(curl -fsSL --retry 2 --retry-delay 5 --retry-max-time 60 \
    "https://api.github.com/repos/LibRaw/LibRaw/releases/latest" | jq -r '.tag_name')
fi

set -euo pipefail

: "${DEB_HOST_MULTIARCH:=$(uname -m)-linux-gnu}"
: "${DEB_HOST_ARCH:=$(dpkg --print-architecture)}"
CACHE_DIR="${BUILD_CACHE_DIR:-/build-cache}/LibRaw-${LIBRAW_VERSION}"
CACHE_MARKER="${CACHE_DIR}/LibRaw-${LIBRAW_VERSION}-complete"

# Check if this specific version is already built and cached
if [[ -f "$CACHE_MARKER" ]] && [[ -d "${CACHE_DIR}/output" ]]; then
  echo "LibRaw ${LIBRAW_VERSION} found in cache, reusing..."
  mkdir -p /output
  cp -ra "${CACHE_DIR}/output/"* /output/
  exit 0
fi

echo "Building LibRaw ${LIBRAW_VERSION} (cache miss)..."

echo Compiler: "${DEB_HOST_MULTIARCH}" Arch: "${DEB_HOST_ARCH}"

apt-get install -y --no-install-recommends \
  libjpeg62-turbo-dev:"${DEB_HOST_ARCH}" \
  liblcms2-dev:"${DEB_HOST_ARCH}" \
  zlib1g-dev:"${DEB_HOST_ARCH}"

URL="https://api.github.com/repos/LibRaw/LibRaw/tarball/${LIBRAW_VERSION}"
echo download libraw from "$URL"
curl -fsSL --retry 2 --retry-delay 5 --retry-max-time 60 -o ./libraw.tar.gz \
  ${GITHUB_TOKEN:+-H "Authorization: Bearer ${GITHUB_TOKEN}"} "$URL"

tar xfv ./libraw.tar.gz
cd LibRaw-*
autoreconf --install
./configure \
  --disable-option-checking \
  --disable-silent-rules \
  --disable-maintainer-mode \
  --disable-dependency-tracking \
  --host="${DEB_HOST_MULTIARCH}"
make
make install
cd ..

mkdir -p /output/bin /output/lib /output/include /output/pkgconfig
cp -a /usr/local/bin/raw* /output/bin/
cp -a /usr/local/lib/libraw_r* /output/lib/
cp -a /usr/local/lib/pkgconfig/libraw* /output/pkgconfig/
cp -a /usr/local/include/libraw /output/include/
file /usr/local/lib/libraw_r.so*

# After successful build, cache the results
echo "Caching LibRaw ${LIBRAW_VERSION} build results..."
mkdir -p "${CACHE_DIR}/output"
cp -ra /output/* "${CACHE_DIR}/output/"
touch "$CACHE_MARKER"

echo "LibRaw ${LIBRAW_VERSION} build complete and cached"
