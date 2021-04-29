#!/bin/bash
set -e

if [ "$TARGETPLATFORM" == "linux/amd64" ]; then
  ALPINE_ARCH='x86_64'
  COMPILER='x86_64-linux-musl-cross'
elif [ "$TARGETPLATFORM" == "linux/arm64" ]; then
  ALPINE_ARCH='aarch64'
  COMPILER='aarch64-linux-musl-cross'
elif [ "$TARGETPLATFORM" == "linux/arm/v6" ] || [ "$TARGETPLATFORM" == "linux/arm/v7" ]; then
  ALPINE_ARCH='armhf'
  COMPILER='arm-linux-musleabihf-cross'
else
  echo "TARGET PLATFORM NOT SUPPORTED: '$TARGETPLATFORM'"; exit 1
fi


apk add --no-cache go curl rsync

# Install G++/GCC cross compilers
curl https://more.musl.cc/x86_64-linux-musl/${COMPILER}.tgz | tar -xz -C /

apk add --arch ${ALPINE_ARCH} --root /${COMPILER} --initdb --allow-untrusted --no-cache \
  -X http://dl-cdn.alpinelinux.org/alpine/edge/community -X http://dl-cdn.alpinelinux.org/alpine/edge/main -X http://dl-cdn.alpinelinux.org/alpine/edge/testing \
  dlib blas-dev openblas-dev lapack-dev jpeg-dev libpng-dev libheif-dev libc-dev

rsync -av /${COMPILER}/lib /
