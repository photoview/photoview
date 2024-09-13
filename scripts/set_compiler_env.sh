#!/bin/sh
set -eu

# Script to configure environment variables for  compiler to cross compilation

: ${TARGETPLATFORM=linux/`dpkg --print-architecture`}

TARGETOS="$(echo $TARGETPLATFORM | cut -d"/" -f1)"
TARGETARCH="$(echo $TARGETPLATFORM | cut -d"/" -f2)"
TARGETVARIANT="$(echo $TARGETPLATFORM | cut -d"/" -f3)"

DEBIAN_ARCH=$TARGETARCH
if [ "$TARGETARCH" = "arm" ]
then
  DEBIAN_ARCH=armhf
  if [ "$TARGETVARIANT" = "v5" ]
  then
    DEBIAN_ARCH=armel
  fi
fi

dpkg --add-architecture $DEBIAN_ARCH
apt-get update
apt-get install -y git curl crossbuild-essential-${DEBIAN_ARCH} libc-dev:${DEBIAN_ARCH} autoconf automake libtool m4 pkg-config cmake

dpkg-architecture -a $DEBIAN_ARCH >/env
export $(cat /env)

CGO_ENABLED="1"
GOOS="$TARGETOS"
GOARCH="$TARGETARCH"

GOARM="7"
if [ "$TARGETARCH" = "arm" && ! -z "$TARGETVARIANT" ]; then
  case "$TARGETVARIANT" in
  "v5")
    export GOARM="5"
    ;;
  "v6")
    export GOARM="6"
    ;;
  esac
fi

echo CGO_ENABLED="${CGO_ENABLED}" >>/env
echo GOOS="${GOOS}" >>/env
echo GOARCH="${GOARCH}" >>/env
echo GOARM="${GOARM}" >>/env
echo AR="${DEB_HOST_MULTIARCH}-ar" >>/env
echo CC="${DEB_HOST_MULTIARCH}-gcc" >>/env
echo CXX="${DEB_HOST_MULTIARCH}-g++" >>/env
echo PKG_CONFIG_PATH="/usr/lib/${DEB_HOST_MULTIARCH}/pkgconfig/" >>/env

cat /env
