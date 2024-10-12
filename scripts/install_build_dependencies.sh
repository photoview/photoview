#!/bin/sh
set -eu

: ${DEB_HOST_ARCH=`dpkg --print-architecture`}
echo Arch: ${DEB_HOST_ARCH}

# Install go-heif dependencies
apt-get install -y libdav1d-dev:${DEB_HOST_ARCH} libde265-dev:${DEB_HOST_ARCH} libjpeg62-turbo-dev:${DEB_HOST_ARCH} libpng-dev:${DEB_HOST_ARCH} libnuma-dev:${DEB_HOST_ARCH} zlib1g-dev:${DEB_HOST_ARCH}

# Install go-face dependencies
apt-get install -y \
  libdlib-dev:${DEB_HOST_ARCH} libblas-dev:${DEB_HOST_ARCH} libatlas-base-dev:${DEB_HOST_ARCH} liblapack-dev:${DEB_HOST_ARCH} libjpeg62-turbo-dev:${DEB_HOST_ARCH}

# Install tools for development
apt-get install -y reflex sqlite3
