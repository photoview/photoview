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
  clang \
  git \
  llvm \
  python3-jsonschema \
  libxml2-utils \
  intltool \
  iso-codes \
  xsltproc \
  zlib1g \
  libavif-dev \
  libcurl4-openssl-dev \
  libexiv2-dev \
  libgmic-dev \
  libgraphicsmagick1-dev \
  libgtk-3-dev \
  libjpeg-dev \
  libjson-glib-dev \
  libjxl-dev \
  liblcms2-dev \
  liblensfun-dev \
  libopenexr-dev \
  libopenjp2-7-dev \
  libpng-dev \
  libpugixml-dev \
  librsvg2-dev \
  libsqlite3-dev \
  libtiff-dev \
  libwebp-dev

URL="https://github.com/darktable-org/darktable.git"
echo download Darktable repo from "$URL"
git clone --depth 1 --single-branch --branch ${DARKTABLE_VERSION} ${URL} darktable || true
cd darktable
git config submodule.src/tests/integration.update none
git submodule update --init --recursive --depth 1 --recommend-shallow --single-branch

FEATURES=" \
  --disable-kwallet \
  --disable-libsecret \
  --disable-lua \
  --disable-mac_integration \
  --disable-map \
  --disable-unity \
  --disable-camera \
  --disable-colord \
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
  ${FEATURES}

mkdir -p /output/opt
cp -a /opt/darktable /output/opt/darktable
file /output/opt/darktable/bin/darktable-cli

# After successful build, cache the results
echo "Caching Darktable ${DARKTABLE_VERSION} build results..."
mkdir -p "${CACHE_DIR}/output"
cp -ra /output/* "${CACHE_DIR}/output/"
touch "$CACHE_MARKER"

echo "Darktable ${DARKTABLE_VERSION} build complete and cached"
