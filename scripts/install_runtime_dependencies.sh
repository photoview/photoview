#!/bin/bash
set -euo pipefail

: "${DEB_HOST_ARCH=`dpkg --print-architecture`}"
echo "Arch: ${DEB_HOST_ARCH}"

if [ "${DEB_HOST_ARCH}" != "$(dpkg --print-architecture)" ]; then
  echo "No need to install runtime dependencies in the cross-build environment, since it can't be run."
  exit 0
fi

LIBS=(
  # compressing static files for better performance
  gzip
  brotli
  zstd

  # health check
  curl

  # exiftool
  libimage-exiftool-perl

  # libraw dependencies
  libgomp1
  libjpeg62-turbo
  liblcms2-2
  zlib1g

  # ImageMagick dependencies
  libgomp1
  libbz2-1.0
  libdjvulibre21
  libheif-plugins-all
  libheif1
  libjbig0
  libjpeg62-turbo
  libjxl0.11
  liblcms2-2
  liblzma5
  libopenexr-3-1-30
  libopenjp2-7
  libpng16-16t64
  libtiff6
  libwmf-0.2-7
  libwebpmux3
  libwebpdemux2
  libwebp7
  libxml2
  libzip5
  libzstd1
  zlib1g

  # go-face dependencies
  libblas3
  libdlib19.2
  libjpeg62-turbo
  liblapack3
)

apt-get update
apt-get install -y --no-install-recommends "${LIBS[@]}"
