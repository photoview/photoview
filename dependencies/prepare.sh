#!/bin/bash
set -euo pipefail

: "${TARGETPLATFORM:=linux/$(dpkg --print-architecture)}"
: "${DEB_HOST_MULTIARCH:=$(uname -m)-linux-gnu}"

TARGETARCH="$(echo "$TARGETPLATFORM" | cut -d"/" -f2)"

DEBIAN_ARCH=$TARGETARCH
if [ "$TARGETARCH" = "arm" ]; then
  DEBIAN_ARCH=armel
fi

dpkg --add-architecture "$DEBIAN_ARCH"
apt-get update
apt-get install -y --no-install-recommends \
  curl \
  jq \
  ca-certificates \
  crossbuild-essential-"${DEBIAN_ARCH}" \
  libc-dev:"${DEBIAN_ARCH}" \
  autoconf \
  automake \
  libtool \
  m4 \
  pkg-config \
  cmake

dpkg-architecture -a "$DEBIAN_ARCH" >/env
echo "PKG_CONFIG_PATH=/usr/lib/${DEB_HOST_MULTIARCH}/pkgconfig" >>/env
# shellcheck disable=SC2046
export $(cat /env)
cat /env
