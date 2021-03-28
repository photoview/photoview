#!/bin/sh

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

# Install Golang
apt-get install -y ca-certificates golang

# Install G++/GCC cross compilers
apt-get install -y \
  g++-aarch64-linux-gnu \
  libc6-dev-arm64-cross \
  g++-arm-linux-gnueabihf \
  libc6-dev-armhf-cross

# Install go-face dependencies
apt-get install -y \
  libdlib-dev:$DEBIAN_ARCH \
  libblas-dev:$DEBIAN_ARCH \
  liblapack-dev:$DEBIAN_ARCH \
  libjpeg62-turbo-dev:$DEBIAN_ARCH \

# Install libheif for HEIF media decoding
apt-get install -y \
  libheif-dev:$DEBIAN_ARCH

# Cleanup
apt-get clean
rm -rf /var/lib/apt/lists/*
