#!/bin/sh
set -eu

: ${DEB_HOST_ARCH=`dpkg --print-architecture`}
echo Arch: ${DEB_HOST_ARCH}
apt-get update

# Install libheif dependencies
apt-get install -y --no-install-recommends libdav1d-dev:${DEB_HOST_ARCH} libde265-dev:${DEB_HOST_ARCH} libjpeg62-turbo-dev:${DEB_HOST_ARCH} libopenh264-dev:${DEB_HOST_ARCH} libpng-dev:${DEB_HOST_ARCH} libnuma-dev:${DEB_HOST_ARCH} zlib1g-dev:${DEB_HOST_ARCH}

# Install libraw dependencies
apt-get install -y --no-install-recommends libjpeg62-turbo-dev:${DEB_HOST_ARCH} liblcms2-dev:${DEB_HOST_ARCH} zlib1g-dev:${DEB_HOST_ARCH}

# Install ImageMagick dependencies
apt-get install -y --no-install-recommends libjxl-dev:${DEB_HOST_ARCH} liblcms2-dev:${DEB_HOST_ARCH} liblqr-1-0-dev:${DEB_HOST_ARCH} libdjvulibre-dev:${DEB_HOST_ARCH} libjpeg62-turbo-dev:${DEB_HOST_ARCH} libopenjp2-7-dev:${DEB_HOST_ARCH} libopenexr-dev:${DEB_HOST_ARCH} libpng-dev:${DEB_HOST_ARCH} libtiff-dev:${DEB_HOST_ARCH} libwebp-dev:${DEB_HOST_ARCH} libxml2-dev:${DEB_HOST_ARCH} libfftw3-dev:${DEB_HOST_ARCH} zlib1g-dev:${DEB_HOST_ARCH} liblzma-dev:${DEB_HOST_ARCH} libbz2-dev:${DEB_HOST_ARCH}

# Install go-face dependencies
apt-get install -y --no-install-recommends libdlib-dev:${DEB_HOST_ARCH} libblas-dev:${DEB_HOST_ARCH} libatlas-base-dev:${DEB_HOST_ARCH} liblapack-dev:${DEB_HOST_ARCH} libjpeg62-turbo-dev:${DEB_HOST_ARCH}

# Install gomagic dependencies
apt-get install -y --no-install-recommends libmagic-dev:${DEB_HOST_ARCH}

# Install tools for development
apt-get install -y --no-install-recommends reflex sqlite3
