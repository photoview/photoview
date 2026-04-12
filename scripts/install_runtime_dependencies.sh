#!/bin/sh
set -eu

: ${DEB_HOST_ARCH=`dpkg --print-architecture`}
echo Arch: ${DEB_HOST_ARCH}

if [ ${DEB_HOST_ARCH} != $(dpkg --print-architecture) ]; then
  echo "No need to install runtime dependencies in the cross-build environment, since it can't be run."
  exit 0
fi

apt-get update

# exiftool and health check
apt-get install -y --no-install-recommends curl libimage-exiftool-perl

# libraw dependencies
apt-get install -y --no-install-recommends \
  libgomp1 \
  libjpeg62-turbo \
  liblcms2-2 \
  zlib1g

# ImageMagick dependencies
apt-get install -y --no-install-recommends \
  libgomp1 \
  libbz2-1.0 \
  libdjvulibre21 \
  libheif1 \
  libjbig0 \
  libjpeg62-turbo \
  libjxl0.11 \
  liblcms2-2 \
  liblzma5 \
  libopenexr-3-1-30 \
  libopenjp2-7 \
  libpng16-16t64 \
  libtiff6 \
  libwmf-0.2-7 \
  libwebpmux3 \
  libwebpdemux2 \
  libwebp7 \
  libxml2 \
  libzip5 \
  libzstd1 \
  zlib1g

# go-face dependencies
apt-get install -y --no-install-recommends \
  libblas3 \
  libdlib19.2 \
  libjpeg62-turbo \
  liblapack3

# libheif dependencies
apt-get install -y --no-install-recommends \
  libdav1d7 \
  librav1e0.7 \
  libde265-0 \
  libx265-215 \
  libjpeg62-turbo \
  libopenh264-8 \
  libpng16-16t64 \
  libnuma1 \
  zlib1g
