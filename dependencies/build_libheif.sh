#!/bin/bash
set -euo pipefail

: "${DEB_HOST_MULTIARCH:=$(uname -m)-linux-gnu}"
: "${DEB_HOST_ARCH:=$(dpkg --print-architecture)}"
CACHE_DIR="${BUILD_CACHE_DIR:-/build-cache}/libheif-${LIBHEIF_VERSION}"
CACHE_MARKER="${CACHE_DIR}/libheif-${LIBHEIF_VERSION}-complete"

# Check if this specific version is already built and cached
if [[ -f "$CACHE_MARKER" ]] && [[ -d "${CACHE_DIR}/output" ]]; then
  echo "libheif ${LIBHEIF_VERSION} found in cache, reusing..."
  mkdir -p /output
  cp -ra "${CACHE_DIR}/output/"* /output/
  exit 0
fi

echo "Building libheif ${LIBHEIF_VERSION} (cache miss)..."

echo Compiler: "${DEB_HOST_MULTIARCH}" Arch: "${DEB_HOST_ARCH}"

apt-get install -y --no-install-recommends \
  libdav1d-dev:"${DEB_HOST_ARCH}" \
  libde265-dev:"${DEB_HOST_ARCH}" \
  libjpeg62-turbo-dev:"${DEB_HOST_ARCH}" \
  libopenh264-dev:"${DEB_HOST_ARCH}" \
  libpng-dev:"${DEB_HOST_ARCH}" \
  libnuma-dev:"${DEB_HOST_ARCH}" \
  zlib1g-dev:"${DEB_HOST_ARCH}"

URL="https://api.github.com/repos/strukturag/libheif/tarball/${LIBHEIF_VERSION}"
echo download libheif from "$URL"
curl -fsSL --retry 2 --retry-delay 5 --retry-max-time 60 -o ./libheif.tar.gz \
  ${GITHUB_TOKEN:+-H "Authorization: Bearer ${GITHUB_TOKEN}"} "$URL"

tar xfv ./libheif.tar.gz
cd ./*-libheif-*
cmake \
  --preset=release \
  -DCMAKE_SYSTEM_PROCESSOR="${DEB_HOST_ARCH}" \
  -DCMAKE_C_COMPILER="${DEB_HOST_MULTIARCH}"-gcc \
  -DCMAKE_CXX_COMPILER="${DEB_HOST_MULTIARCH}"-g++ \
  -DPKG_CONFIG_EXECUTABLE="${DEB_HOST_MULTIARCH}"-pkg-config \
  -DCMAKE_LIBRARY_ARCHITECTURE="${DEB_HOST_MULTIARCH}" \
  -DENABLE_PLUGIN_LOADING=OFF \
  -DWITH_GDK_PIXBUF=OFF .
make
make install
cd ..

mkdir -p /output/bin /output/lib /output/include /output/pkgconfig
cp -a /usr/local/bin/heif-* /output/bin/
cp -a /usr/local/lib/libheif* /output/lib/
cp -a /usr/local/lib/pkgconfig/libheif* /output/pkgconfig/
cp -a /usr/local/include/libheif /output/include/
file /usr/local/lib/libheif.so*

# After successful build, cache the results
echo "Caching libheif ${LIBHEIF_VERSION} build results..."
mkdir -p "${CACHE_DIR}/output"
cp -ra /output/* "${CACHE_DIR}/output/"
touch "$CACHE_MARKER"

echo "libheif ${LIBHEIF_VERSION} build complete and cached"
