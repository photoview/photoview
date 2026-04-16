#!/bin/bash
set -euo pipefail

# Configure environment for cross-compiling.

: "${TARGETPLATFORM:=linux/$(dpkg --print-architecture)}"

echo "Target platform: ${TARGETPLATFORM}"

TARGETOS="$(echo ${TARGETPLATFORM} | cut -d"/" -f1)"
TARGETARCH="$(echo ${TARGETPLATFORM} | cut -d"/" -f2)"
TARGETVARIANT="$(echo ${TARGETPLATFORM} | cut -d"/" -f3)"

DEBIAN_ARCH="${TARGETARCH}"
if [[ "${TARGETARCH}" != "amd64" && "${TARGETARCH}" != "arm64" ]]; then
  echo "Warning: ${TARGETPLATFORM} is NOT supported. Just compile with the best efforts."

  # best efforts
  DEBIAN_ARCH="armel"
  if [ "${TARGETVARIANT}" = "v7" ]
  then
    DEBIAN_ARCH="armhf"
  fi
fi

CGO_ENABLED="1"
GOOS="${TARGETOS}"
GOARCH="${TARGETARCH}"
GOARM=""
# best efforts
if [ "${TARGETARCH}" = "arm" ] && [ ! -z "${TARGETVARIANT}" ]; then
  GOARM="7"
  case "${TARGETVARIANT}" in
  "v5")
    GOARM="5"
    ;;
  "v6")
    GOARM="6"
    ;;
  esac
fi

LIBS=(
  autoconf
  automake
  cmake
  curl
  dpkg-dev
  git
  libtool
  m4
  pkg-config

  # for crossbuild
  "crossbuild-essential-${DEBIAN_ARCH}"
  "libc-dev:${DEBIAN_ARCH}"
)

dpkg --add-architecture "${DEBIAN_ARCH}"
apt-get update
apt-get install -y --no-install-recommends "${LIBS[@]}"

dpkg-architecture -a "${DEBIAN_ARCH}" >/env
set -a
source /env
set +a

echo CGO_ENABLED="${CGO_ENABLED}" >>/env
echo GOOS="${GOOS}" >>/env
echo GOARCH="${GOARCH}" >>/env
echo GOARM="${GOARM}" >>/env
echo AR="${DEB_HOST_MULTIARCH}-ar" >>/env
echo CC="${DEB_HOST_MULTIARCH}-gcc" >>/env
echo CXX="${DEB_HOST_MULTIARCH}-g++" >>/env
echo PKG_CONFIG_PATH="/usr/lib/${DEB_HOST_MULTIARCH}/pkgconfig/" >>/env

cat /env
