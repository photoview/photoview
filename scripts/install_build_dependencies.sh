#!/bin/bash
set -e

if [ "$TARGETPLATFORM" == "linux/arm64" ]; then
  dpkg --add-architecture arm64
  DEBIAN_ARCH='arm64'
elif [ "$TARGETPLATFORM" == "linux/arm/v6" ] || [ "$TARGETPLATFORM" == "linux/arm/v7" ]; then
  dpkg --add-architecture armhf
  DEBIAN_ARCH='armhf'
else
  dpkg --add-architecture amd64
  DEBIAN_ARCH='amd64'
fi

apt update

# Install G++/GCC cross compilers
if [ "$DEBIAN_ARCH" == "arm64" ]; then
  apt install -y \
    g++-aarch64-linux-gnu \
    libc6-dev-arm64-cross
elif [ "$DEBIAN_ARCH" == "armhf" ]; then
  apt install -y \
    g++-arm-linux-gnueabihf \
    libc6-dev-armhf-cross
else
  apt install -y \
    g++-x86-64-linux-gnu \
    libc6-dev-amd64-cross
fi

# Install go-face dependencies and libheif for HEIF media decoding
apt install -y \
  libdlib-dev:${DEBIAN_ARCH} \
  libblas-dev:${DEBIAN_ARCH} \
  libatlas-base-dev:${DEBIAN_ARCH} \
  liblapack-dev:${DEBIAN_ARCH} \
  libjpeg-dev:${DEBIAN_ARCH} \
  libheif-dev:${DEBIAN_ARCH}

# Cleanup
apt clean
rm -rf /var/lib/apt/lists/*
