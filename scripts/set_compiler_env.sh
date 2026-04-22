#!/bin/bash
set -euo pipefail

# Configure environment for cross-compiling.

: "${TARGETPLATFORM:=linux/$(dpkg --print-architecture)}"

echo "Target platform: ${TARGETPLATFORM}"

TARGETOS="$(echo "${TARGETPLATFORM}" | cut -d'/' -f1)"
TARGETARCH="$(echo "${TARGETPLATFORM}" | cut -d'/' -f2)"
TARGETVARIANT="$(echo "${TARGETPLATFORM}" | cut -d'/' -f3)"

DEBIAN_ARCH="${TARGETARCH}"
if [[ "${TARGETARCH}" != "amd64" && "${TARGETARCH}" != "arm64" ]]; then
  echo "Warning: ${TARGETPLATFORM} is NOT supported. Just compile with the best efforts."

  # best efforts
  if [[ "${TARGETARCH}" = "arm" ]]; then
    if [[ "${TARGETVARIANT}" = "v7" ]]; then
      DEBIAN_ARCH="armhf"
    else
      DEBIAN_ARCH="armel"
    fi
  fi
fi

CGO_ENABLED="1"
GOOS="${TARGETOS}"
GOARCH="${TARGETARCH}"
GOARM=""
# best efforts
if [[ "${TARGETARCH}" = "arm" && -n "${TARGETVARIANT}" ]]; then
  case "${TARGETVARIANT}" in
  "v5")
    GOARM="5"
    ;;
  "v6")
    GOARM="6"
    ;;
  "v7")
    GOARM="7"
    ;;
  *)
    GOARM="7"
    DEBIAN_ARCH="armhf"
    echo "Warning: unexpected ARM variant ${TARGETVARIANT}; defaulting GOARM to 7."
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
# shellcheck disable=SC1091
source /env
set +a

{
  echo CGO_ENABLED="${CGO_ENABLED}"
  echo GOOS="${GOOS}"
  echo GOARCH="${GOARCH}"
  echo GOARM="${GOARM}"
  echo AR="${DEB_HOST_MULTIARCH}-ar"
  echo CC="${DEB_HOST_MULTIARCH}-gcc"
  echo CXX="${DEB_HOST_MULTIARCH}-g++"
  echo PKG_CONFIG_PATH="/usr/lib/${DEB_HOST_MULTIARCH}/pkgconfig/"
} >> /env

cat /env
