#!/bin/bash

# Fallback to the latest version if IMAGEMAGICK_VERSION is not set
if [[ -z "${IMAGEMAGICK_VERSION}" ]]; then
  echo "WARN: ImageMagick version is empty, most likely the script runs not on CI."
  echo "Fetching the latest version from ImageMagick repo..."
  IMAGEMAGICK_VERSION=$(curl -fsSL --retry 2 --retry-delay 5 --retry-max-time 60 \
    "https://api.github.com/repos/ImageMagick/ImageMagick/releases/latest" | jq -r '.tag_name')
fi

set -euo pipefail

: "${DEB_TARGET_MULTIARCH:=$(uname -m)-linux-gnu}"
: "${DEB_TARGET_ARCH:=$(dpkg --print-architecture)}"
CACHE_DIR="${BUILD_CACHE_DIR:-/build-cache}/ImageMagick-${IMAGEMAGICK_VERSION}"
CACHE_MARKER="${CACHE_DIR}/ImageMagick-${IMAGEMAGICK_VERSION}-complete"

# Check if this specific version is already built and cached
if [[ -f "$CACHE_MARKER" ]] && [[ -d "${CACHE_DIR}/output" ]]; then
  echo "ImageMagick ${IMAGEMAGICK_VERSION} found in cache, reusing..."
  mkdir -p /output
  cp -ra "${CACHE_DIR}/output/"* /output/
  exit 0
fi

echo "Building ImageMagick ${IMAGEMAGICK_VERSION} (cache miss)..."

echo Compiler: "${DEB_TARGET_MULTIARCH}" Arch: "${DEB_TARGET_ARCH}"

apt-get install -y \
  libjxl-dev:"${DEB_TARGET_ARCH}" \
  liblcms2-dev:"${DEB_TARGET_ARCH}" \
  liblqr-1-0-dev:"${DEB_TARGET_ARCH}" \
  libdjvulibre-dev:"${DEB_TARGET_ARCH}" \
  libjpeg62-turbo-dev:"${DEB_TARGET_ARCH}" \
  libopenjp2-7-dev:"${DEB_TARGET_ARCH}" \
  libopenexr-dev:"${DEB_TARGET_ARCH}" \
  libpng-dev:"${DEB_TARGET_ARCH}" \
  libtiff-dev:"${DEB_TARGET_ARCH}" \
  libwebp-dev:"${DEB_TARGET_ARCH}" \
  libxml2-dev:"${DEB_TARGET_ARCH}" \
  libfftw3-dev:"${DEB_TARGET_ARCH}" \
  zlib1g-dev:"${DEB_TARGET_ARCH}" \
  liblzma-dev:"${DEB_TARGET_ARCH}" \
  libbz2-dev:"${DEB_TARGET_ARCH}"

URL="https://api.github.com/repos/ImageMagick/ImageMagick/tarball/${IMAGEMAGICK_VERSION}"
echo download ImageMagick from "$URL"
curl -fsSL --retry 2 --retry-delay 5 --retry-max-time 60 -o ./magick.tar.gz \
  ${GITHUB_TOKEN:+-H "Authorization: Bearer ${GITHUB_TOKEN}"} "$URL"

FEATURES="--with-heic --with-jpeg --with-png --with-raw --with-tiff --with-webp"

tar xfv ./magick.tar.gz
cd ImageMagick-*
./configure \
  --enable-64bit-channel-masks \
  --enable-static --enable-shared --enable-delegate-build \
  --without-x --without-magick-plus-plus \
  --without-perl --disable-doc \
  --host="${DEB_TARGET_MULTIARCH}" \
  ${FEATURES}

# Ensure that features are enabled
for feature in ${FEATURES}
do
  grep -- ${feature}'.*yes$' config.log || (echo "Can't enable feature ${feature}"; false)
done

make
make install
cd ..

mkdir -p /output/bin /output/etc /output/lib /output/include /output/pkgconfig
cp -a /usr/local/bin/magick /output/bin/
cp -a /usr/local/etc/ImageMagick-7 /output/etc/
cp -a /usr/local/lib/ImageMagick-* /output/lib/
cp -a /usr/local/lib/libMagickCore-* /output/lib/
cp -a /usr/local/lib/libMagickWand-* /output/lib/
cp -a /usr/local/include/ImageMagick-7 /output/include/
cp -a /usr/local/lib/pkgconfig/ImageMagick*.pc /output/pkgconfig/
cp -a /usr/local/lib/pkgconfig/MagickCore*.pc /output/pkgconfig/
cp -a /usr/local/lib/pkgconfig/MagickWand*.pc /output/pkgconfig/
file /output/bin/magick

# After successful build, cache the results
echo "Caching ImageMagick ${IMAGEMAGICK_VERSION} build results..."
mkdir -p "${CACHE_DIR}/output"
cp -ra /output/* "${CACHE_DIR}/output/"
touch "$CACHE_MARKER"

echo "ImageMagick ${IMAGEMAGICK_VERSION} build complete and cached"
