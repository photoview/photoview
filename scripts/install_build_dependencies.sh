#!/bin/sh
set -eu

: ${DEB_HOST_ARCH=`dpkg --print-architecture`}
echo Arch: ${DEB_HOST_ARCH}

# libraw dependencies
apt-get install -y zlib1g-dev:${DEB_HOST_ARCH} liblcms2-dev:${DEB_HOST_ARCH} libjpeg-dev:${DEB_HOST_ARCH}

# libheif dependencies
apt-get install -y libaom-dev:${DEB_HOST_ARCH} libavcodec-dev:${DEB_HOST_ARCH} libdav1d-dev:${DEB_HOST_ARCH} libde265-dev:${DEB_HOST_ARCH} libgdk-pixbuf-2.0-dev:${DEB_HOST_ARCH} libjpeg-dev:${DEB_HOST_ARCH} libopenjp2-7-dev:${DEB_HOST_ARCH} libpng-dev:${DEB_HOST_ARCH} librav1e-dev:${DEB_HOST_ARCH} libsvtav1enc-dev:${DEB_HOST_ARCH} libx265-dev:${DEB_HOST_ARCH}

# ImageMagick dependencies
apt-get install -y libjxl-dev:${DEB_HOST_ARCH} libfftw3-dev:${DEB_HOST_ARCH} liblcms2-dev:${DEB_HOST_ARCH} liblqr-1-0-dev:${DEB_HOST_ARCH} zlib1g-dev:${DEB_HOST_ARCH} liblzma-dev:${DEB_HOST_ARCH} libbz2-dev:${DEB_HOST_ARCH} libdjvulibre-dev:${DEB_HOST_ARCH} libexif-dev:${DEB_HOST_ARCH} libjpeg-dev:${DEB_HOST_ARCH} libopenjp2-7-dev:${DEB_HOST_ARCH} libopenexr-dev:${DEB_HOST_ARCH} libpng-dev:${DEB_HOST_ARCH} libtiff-dev:${DEB_HOST_ARCH} libwmf-dev:${DEB_HOST_ARCH} libwebp-dev:${DEB_HOST_ARCH} libxml2-dev:${DEB_HOST_ARCH}

# Install go-face dependencies and libheif for HEIF media decoding
apt-get install -y \
  libdlib-dev:${DEB_HOST_ARCH} libblas-dev:${DEB_HOST_ARCH} libatlas-base-dev:${DEB_HOST_ARCH} liblapack-dev:${DEB_HOST_ARCH} libjpeg-dev:${DEB_HOST_ARCH}

# Install tools for development
apt-get install -y reflex sqlite3
