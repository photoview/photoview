#!/bin/bash
set -euo pipefail

: "${TARGETPLATFORM:=linux/$(dpkg --print-architecture)}"

TARGETARCH="$(echo "$TARGETPLATFORM" | cut -d"/" -f2)"

DEBIAN_ARCH=$TARGETARCH
if [[ "$TARGETARCH" = "arm" ]]; then
  DEBIAN_ARCH=armel
fi

dpkg --add-architecture "$DEBIAN_ARCH"
apt-get update
apt-get install -y \
  autoconf \
  automake \
  ca-certificates \
  clang \
  cmake \
  curl \
  dpkg-dev \
  git \
  intltool \
  iso-codes \
  jq \
  llvm \
  libtool \
  libxml2-utils \
  m4 \
  pkg-config \
  python3-jsonschema \
  xsltproc \
  crossbuild-essential-"${DEBIAN_ARCH}" \
  libc-dev:"${DEBIAN_ARCH}"

dpkg-architecture -a "$DEBIAN_ARCH" >/env
echo "PKG_CONFIG=$(which pkg-config)" >>/env
echo "PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:/usr/lib/$(dpkg-architecture -a "$DEBIAN_ARCH" -qDEB_HOST_MULTIARCH)/pkgconfig" >>/env
cat /env
