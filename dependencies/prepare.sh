#!/bin/sh
set -eu

: ${TARGETPLATFORM=linux/`dpkg --print-architecture`}

TARGETOS="$(echo $TARGETPLATFORM | cut -d"/" -f1)"
TARGETARCH="$(echo $TARGETPLATFORM | cut -d"/" -f2)"
TARGETVARIANT="$(echo $TARGETPLATFORM | cut -d"/" -f3)"

DEBIAN_ARCH=$TARGETARCH
if [ "$TARGETARCH" = "arm" ]
then
  DEBIAN_ARCH=armel
  if [ "$TARGETVARIANT" = "v7" ]
  then
    DEBIAN_ARCH=armhf
  fi
fi

dpkg --add-architecture $DEBIAN_ARCH
apt-get update
apt-get install -y curl crossbuild-essential-${DEBIAN_ARCH} libc-dev:${DEBIAN_ARCH} autoconf automake libtool m4 pkg-config cmake

dpkg-architecture -a $DEBIAN_ARCH >/env
cat /env
