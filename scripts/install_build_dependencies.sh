#!/bin/sh
set -eu

: ${DEB_HOST_ARCH=`dpkg --print-architecture`}
echo Arch: ${DEB_HOST_ARCH}

# Install go-face dependencies
apt-get install -y \
  libdlib-dev:${DEB_HOST_ARCH} libblas-dev:${DEB_HOST_ARCH} libatlas-base-dev:${DEB_HOST_ARCH} liblapack-dev:${DEB_HOST_ARCH} libjpeg-dev:${DEB_HOST_ARCH}

# Install tools for development
apt-get install -y reflex sqlite3
