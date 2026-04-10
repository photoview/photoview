#!/bin/sh
set -eu

: ${DEB_HOST_ARCH=`dpkg --print-architecture`}
echo Target Arch: ${DEB_HOST_ARCH}
echo Env Arch: $(dpkg --print-architecture)

apt-get update

# Install go-face dependencies
apt-get install -y --no-install-recommends libdlib-dev:${DEB_HOST_ARCH} libblas-dev:${DEB_HOST_ARCH} liblapack-dev:${DEB_HOST_ARCH} libjpeg62-turbo-dev:${DEB_HOST_ARCH}

# Install graphicswand dependencies
apt-get install -y --no-install-recommends libgraphicsmagick1-dev:${DEB_HOST_ARCH}

# Install tools for development
apt-get install -y --no-install-recommends reflex sqlite3
