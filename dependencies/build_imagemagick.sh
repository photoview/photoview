#!/bin/bash

set -euo pipefail

: "${IMAGEMAGICK_VERSION:=}"

# Fallback to the latest version if IMAGEMAGICK_VERSION is not set
if [[ -z "${IMAGEMAGICK_VERSION}" ]]; then
  echo "WARN: ImageMagick version is empty, most likely the script runs not on CI."
  echo "Fetching the latest version from ImageMagick repo..."
  IMAGEMAGICK_VERSION="$(curl -fsSL --retry 2 --retry-delay 5 --retry-max-time 60 \
    "https://api.github.com/repos/ImageMagick/ImageMagick/releases/latest" | jq -r '.tag_name // ""')"
  if [[ -z "${IMAGEMAGICK_VERSION}" ]]; then
    echo "ERROR: Failed to resolve latest ImageMagick tag_name from GitHub API" >&2
    exit 1
  fi
fi

: "${DEB_HOST_ARCH:=$(dpkg --print-architecture)}"
: "${DEB_HOST_GNU_TYPE:=$(dpkg-architecture -a "$DEB_HOST_ARCH" -qDEB_HOST_GNU_TYPE)}"

echo "Compiler: ${DEB_HOST_GNU_TYPE} Arch: ${DEB_HOST_ARCH}"

apt-get install -y --no-install-recommends \
  "libbz2-dev:${DEB_HOST_ARCH}" \
  "libdjvulibre-dev:${DEB_HOST_ARCH}" \
  "libfftw3-dev:${DEB_HOST_ARCH}" \
  "libheif-dev:${DEB_HOST_ARCH}" \
  "libjbig-dev:${DEB_HOST_ARCH}" \
  "libjpeg62-turbo-dev:${DEB_HOST_ARCH}" \
  "libjxl-dev:${DEB_HOST_ARCH}" \
  "liblcms2-dev:${DEB_HOST_ARCH}" \
  "liblzma-dev:${DEB_HOST_ARCH}" \
  "libopenexr-dev:${DEB_HOST_ARCH}" \
  "libopenjp2-7-dev:${DEB_HOST_ARCH}" \
  "libpng-dev:${DEB_HOST_ARCH}" \
  "libtiff-dev:${DEB_HOST_ARCH}" \
  "libwebp-dev:${DEB_HOST_ARCH}" \
  "libwmf-dev:${DEB_HOST_ARCH}" \
  "libxml2-dev:${DEB_HOST_ARCH}" \
  "libzip-dev:${DEB_HOST_ARCH}" \
  "libzstd-dev:${DEB_HOST_ARCH}" \
  "zlib1g-dev:${DEB_HOST_ARCH}"

echo "Building ImageMagick ${IMAGEMAGICK_VERSION}..."

URL="https://api.github.com/repos/ImageMagick/ImageMagick/tarball/${IMAGEMAGICK_VERSION}"
echo download ImageMagick from "$URL"
curl -fsSL --retry 2 --retry-delay 5 --retry-max-time 60 -o ./magick.tar.gz \
  ${GITHUB_TOKEN:+-H "Authorization: Bearer ${GITHUB_TOKEN}"} "$URL"

tar xfv ./magick.tar.gz
cd ImageMagick-*

FEATURES=(
  --with-bzlib
  --with-djvu
  --with-heic
  --with-jbig
  --with-jpeg
  --with-jxl
  --with-lcms
  --with-lzma
  --with-openexr
  --with-openjp2
  --with-png
  --with-raw
  --with-tiff
  --with-webp
  --with-wmf
  --with-xml
  --with-zip
  --with-zstd
)

./configure \
  --enable-64bit-channel-masks \
  --enable-static \
  --enable-shared \
  --enable-delegate-build \
  "${FEATURES[@]}" \
  --without-x \
  --without-magick-plus-plus \
  --without-perl \
  "--host=${DEB_HOST_GNU_TYPE}" \
  --prefix=/usr/local

# Ensure that features are enabled
for feature in "${FEATURES[@]}"
do
  grep -- "${feature}.*yes\$" config.log || (echo "Can't enable feature ${feature}"; false)
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

echo "ImageMagick ${IMAGEMAGICK_VERSION} build complete"
