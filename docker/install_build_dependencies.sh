#!/bin/bash
set -e

if [ "$TARGETPLATFORM" == "linux/arm64" ]; then
  dpkg --add-architecture arm64
  DEBIAN_ARCH='arm64'
elif [ "$TARGETPLATFORM" == "linux/arm/v6" ] || [ "$TARGETPLATFORM" == "linux/arm/v7" ]; then
  dpkg --add-architecture armhf
  DEBIAN_ARCH='armhf'
else
  DEBIAN_ARCH='amd64'
fi

apt-get update

# Install G++/GCC cross compilers
if [ "$DEBIAN_ARCH" == "arm64" ]; then
  apt-get install -y \
    g++-aarch64-linux-gnu \
    libc6-dev-arm64-cross
elif [ "$DEBIAN_ARCH" == "armhf" ]; then
  apt-get install -y \
    g++-arm-linux-gnueabihf \
    libc6-dev-armhf-cross
fi

# Install Golang
apt-get install -y ca-certificates golang

# Install go-face dependencies and libheif for HEIF media decoding
apt-get install -y \
  libdlib-dev:${DEBIAN_ARCH} \
  libblas-dev:${DEBIAN_ARCH} \
  libatlas-base-dev:${DEBIAN_ARCH} \
  liblapack-dev:${DEBIAN_ARCH} \
  libjpeg-dev:${DEBIAN_ARCH} \
  libheif-dev:${DEBIAN_ARCH}

# Cleanup
apt-get clean
rm -rf /var/lib/apt/lists/*
