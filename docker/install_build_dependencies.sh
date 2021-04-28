#!/bin/bash
set -e

if [ "$TARGETPLATFORM" == "linux/arm64" ]; then
  DEBIAN_ARCH='arm64'
elif [ "$TARGETPLATFORM" == "linux/arm/v6" ] || [ "$TARGETPLATFORM" == "linux/arm/v7" ]; then
  DEBIAN_ARCH='armhf'
else
  DEBIAN_ARCH='amd64'
fi

# apt-get update

# Install Golang
# apt-get install -y ca-certificates golang curl

apk add --no-cache go curl rsync

# Install go-face dependencies and libheif for HEIF media decoding
# apt-get install -y \
#   libdlib-dev:${DEBIAN_ARCH} \
#   libblas-dev:${DEBIAN_ARCH} \
#   liblapack-dev:${DEBIAN_ARCH} \
#   libjpeg-dev:${DEBIAN_ARCH} \
#   libheif-dev:${DEBIAN_ARCH}



# Install G++/GCC cross compilers
if [ "$DEBIAN_ARCH" == "arm64" ]; then
  curl https://more.musl.cc/x86_64-linux-musl/aarch64-linux-musl-cross.tgz | tar -xz -C /

  apk add --arch aarch64 --root /aarch64-linux-musl-cross --initdb -X http://dl-cdn.alpinelinux.org/alpine/edge/community -X http://dl-cdn.alpinelinux.org/alpine/edge/main -X http://dl-cdn.alpinelinux.org/alpine/edge/testing --allow-untrusted --no-cache \
    dlib openblas-dev lapack-dev jpeg-dev libheif-dev

  rm /aarch64-linux-musl-cross/bin/ld
  ln /aarch64-linux-musl-cross/bin/aarch64-linux-musl-ld /aarch64-linux-musl-cross/bin/ld


  # rsync -av /aarch64-linux-musl-cross/lib /usr
  # rsync -av /aarch64-linux-musl-cross/include /usr

  # rm /aarch64-linux-musl-cross/usr
  # cp -R /aarch64-linux-musl-cross/lib /aarch64-linux-musl-cross/usr

  # rsync --ignore-errors -av /aarch64-linux-musl-cross/ /

  # apt-get install -y \
  #   g++-aarch64-linux-gnu \
  #   libc6-dev-arm64-cross
elif [ "$DEBIAN_ARCH" == "armhf" ]; then
  apt-get install -y \
    g++-arm-linux-gnueabihf \
    libc6-dev-armhf-cross
fi
