#!/bin/bash

set -euo pipefail

: "${LIBRAW_VERSION:=}"

# Fallback to the latest version if LIBRAW_VERSION is not set
if [[ -z "${LIBRAW_VERSION}" ]]; then
  echo "WARN: LibRaw version is empty, most likely the script runs not on CI."
  echo "Fetching the latest version from LibRaw repo..."
  LIBRAW_VERSION="$(curl -fsSL --retry 2 --retry-delay 5 --retry-max-time 60 \
    "https://api.github.com/repos/LibRaw/LibRaw/releases/latest" | jq -r '.tag_name // ""')"
  if [[ -z "${LIBRAW_VERSION}" ]]; then
    echo "ERROR: Failed to resolve latest libraw tag_name from GitHub API" >&2
    exit 1
  fi
fi

: "${DEB_HOST_ARCH:=$(dpkg --print-architecture)}"
: "${DEB_HOST_GNU_TYPE:=$(dpkg-architecture -a "$DEB_HOST_ARCH" -qDEB_HOST_GNU_TYPE)}"

echo "Compiler: ${DEB_HOST_GNU_TYPE} Arch: ${DEB_HOST_ARCH}"

apt-get install -y --no-install-recommends \
  "libjpeg62-turbo-dev:${DEB_HOST_ARCH}" \
  "liblcms2-dev:${DEB_HOST_ARCH}" \
  "zlib1g-dev:${DEB_HOST_ARCH}"

echo "Building LibRaw ${LIBRAW_VERSION}..."

URL="https://api.github.com/repos/LibRaw/LibRaw/tarball/${LIBRAW_VERSION}"
echo "download libraw from ${URL}"
curl -fsSL --retry 2 --retry-delay 5 --retry-max-time 60 -o ./libraw.tar.gz \
  ${GITHUB_TOKEN:+-H "Authorization: Bearer ${GITHUB_TOKEN}"} "${URL}"

tar xfv ./libraw.tar.gz
cd LibRaw-*
autoreconf --install
./configure \
  --enable-static=yes \
  --enable-shared=yes \
  --enable-openmp \
  --enable-jpeg \
  --enable-zlib \
  --enable-lcms \
  --disable-examples \
  --disable-silent-rules \
  --disable-dependency-tracking \
  --host="${DEB_HOST_GNU_TYPE}" \
  --prefix=/usr/local
make
make install
cd ..

mkdir -p /output/bin /output/lib /output/include /output/pkgconfig
cp -a /usr/local/lib/libraw* /output/lib/
cp -a /usr/local/lib/pkgconfig/libraw* /output/pkgconfig/
cp -a /usr/local/include/libraw /output/include/
file /usr/local/lib/libraw_r.so*

echo "LibRaw ${LIBRAW_VERSION} build complete"
