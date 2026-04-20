#!/bin/bash
set -euo pipefail

: "${DEB_HOST_ARCH:=$(dpkg --print-architecture)}"
echo "Arch: ${DEB_HOST_ARCH}"

LIBS=(
  # libraw dependencies
  "libjpeg62-turbo-dev:${DEB_HOST_ARCH}"
  "liblcms2-dev:${DEB_HOST_ARCH}"
  "zlib1g-dev:${DEB_HOST_ARCH}"

  # ImageMagick dependencies
  "libbz2-dev:${DEB_HOST_ARCH}"
  "libdjvulibre-dev:${DEB_HOST_ARCH}"
  "libfftw3-dev:${DEB_HOST_ARCH}"
  "libheif-dev:${DEB_HOST_ARCH}"
  "libjbig-dev:${DEB_HOST_ARCH}"
  "libjpeg62-turbo-dev:${DEB_HOST_ARCH}"
  "libjxl-dev:${DEB_HOST_ARCH}"
  "liblcms2-dev:${DEB_HOST_ARCH}"
  "liblqr-1-0-dev:${DEB_HOST_ARCH}"
  "liblzma-dev:${DEB_HOST_ARCH}"
  "libopenexr-dev:${DEB_HOST_ARCH}"
  "libopenjp2-7-dev:${DEB_HOST_ARCH}"
  "libpng-dev:${DEB_HOST_ARCH}"
  "libtiff-dev:${DEB_HOST_ARCH}"
  "libwebp-dev:${DEB_HOST_ARCH}"
  "libwmf-dev:${DEB_HOST_ARCH}"
  "libxml2-dev:${DEB_HOST_ARCH}"
  "libzip-dev:${DEB_HOST_ARCH}"
  "libzstd-dev:${DEB_HOST_ARCH}"
  "zlib1g-dev:${DEB_HOST_ARCH}"

  # go-face dependencies
  "libdlib-dev:${DEB_HOST_ARCH}"
  "libblas-dev:${DEB_HOST_ARCH}"
  "liblapack-dev:${DEB_HOST_ARCH}"
  "libjpeg62-turbo-dev:${DEB_HOST_ARCH}"

  # tools for development
  reflex
  sqlite3
)

apt-get update
apt-get install -y --no-install-recommends "${LIBS[@]}"
