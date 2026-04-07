#!/bin/bash

# Fallback to the latest version if DARKTABLE_VERSION is not set
if [[ -z "${DARKTABLE_VERSION}" ]]; then
  echo "WARN: Darktable version is empty, most likely the script runs not on CI."
  echo "Fetching the latest version from Darktable repo..."
  DARKTABLE_VERSION=$(curl -fsSL --retry 2 --retry-delay 5 --retry-max-time 60 \
    "https://api.github.com/repos/darktable-org/darktable/releases/latest" | jq -r '.tag_name')
fi

set -euo pipefail

: "${DEB_HOST_ARCH:=$(dpkg --print-architecture)}"
: "${DEB_HOST_GNU_TYPE:=$(dpkg-architecture -a "$DEB_HOST_ARCH" -qDEB_HOST_GNU_TYPE)}"
CACHE_DIR="${BUILD_CACHE_DIR:-/build-cache}/Darktable-${DARKTABLE_VERSION}"
CACHE_MARKER="${CACHE_DIR}/Darktable-${DARKTABLE_VERSION}-complete"

# Check if this specific version is already built and cached
if [[ -f "$CACHE_MARKER" ]] && [[ -d "${CACHE_DIR}/output" ]]; then
  echo "Darktable ${DARKTABLE_VERSION} found in cache, reusing..."
  mkdir -p /output
  cp -ra "${CACHE_DIR}/output/"* /output/
  exit 0
fi

echo "Building Darktable ${DARKTABLE_VERSION} (cache miss)..."

echo Compiler: "${DEB_HOST_GNU_TYPE}" Arch: "${DEB_HOST_ARCH}"

apt-get install -y \
  libavif-dev:"${DEB_HOST_ARCH}" \
  libcurl4-openssl-dev:"${DEB_HOST_ARCH}" \
  libexiv2-dev:"${DEB_HOST_ARCH}" \
  libgmic-dev:"${DEB_HOST_ARCH}" \
  libgraphicsmagick1-dev:"${DEB_HOST_ARCH}" \
  libgtk-3-dev:"${DEB_HOST_ARCH}" \
  libjpeg-dev:"${DEB_HOST_ARCH}" \
  libjson-glib-dev:"${DEB_HOST_ARCH}" \
  libjxl-dev:"${DEB_HOST_ARCH}" \
  liblcms2-dev:"${DEB_HOST_ARCH}" \
  liblensfun-dev:"${DEB_HOST_ARCH}" \
  libopenexr-dev:"${DEB_HOST_ARCH}" \
  libopenjp2-7-dev:"${DEB_HOST_ARCH}" \
  libpng-dev:"${DEB_HOST_ARCH}" \
  libpotrace-dev:"${DEB_HOST_ARCH}" \
  libpugixml-dev:"${DEB_HOST_ARCH}" \
  librsvg2-dev:"${DEB_HOST_ARCH}" \
  libsqlite3-dev:"${DEB_HOST_ARCH}" \
  libtiff-dev:"${DEB_HOST_ARCH}" \
  libwebp-dev:"${DEB_HOST_ARCH}"

URL="https://github.com/darktable-org/darktable.git"
echo download Darktable repo from "$URL"
git clone --depth 1 --single-branch --branch ${DARKTABLE_VERSION} ${URL} darktable || true
cd darktable
git config submodule.src/tests/integration.update none
git config submodule.src/external/lua-scripts.update none
git submodule update --init --recursive --depth 1 --recommend-shallow --single-branch

FEATURES=" \
  --disable-camera \
  --disable-unity \
  --disable-colord \
  --disable-kwallet \
  --disable-libsecret \
  --disable-lua \
  --disable-mac_integration \
  --disable-map \
  --enable-graphicsmagick \
  --enable-imagemagick \
  --enable-jxl \
  --enable-opencl \
  --enable-openexr \
  --enable-openmp \
  --enable-webp"

./build.sh \
  --prefix "/opt/darktable" \
  --build-type "Release" \
  --install \
  ${FEATURES} \
  -- \
  -DCMAKE_SYSTEM_PROCESSOR="${DEB_HOST_ARCH}" \
  -DCMAKE_C_COMPILER="${DEB_HOST_GNU_TYPE}"-gcc \
  -DCMAKE_CXX_COMPILER="${DEB_HOST_GNU_TYPE}"-g++ \
  -DCMAKE_LIBRARY_ARCHITECTURE="${DEB_HOST_GNU_TYPE}"

mkdir -p /output/opt
cp -a /opt/darktable /output/opt/darktable
file /output/opt/darktable/bin/darktable-cli

# After successful build, cache the results
echo "Caching Darktable ${DARKTABLE_VERSION} build results..."
mkdir -p "${CACHE_DIR}/output"
cp -ra /output/* "${CACHE_DIR}/output/"
touch "$CACHE_MARKER"

echo "Darktable ${DARKTABLE_VERSION} build complete and cached"
