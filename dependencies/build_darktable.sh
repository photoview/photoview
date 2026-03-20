#!/bin/bash

# Fallback to the latest version if DARKTABLE_VERSION is not set
if [[ -z "${DARKTABLE_VERSION}" ]]; then
  echo "WARN: Darktable version is empty, most likely the script runs not on CI."
  echo "Fetching the latest version from Darktable repo..."
  DARKTABLE_VERSION=$(curl -fsSL --retry 2 --retry-delay 5 --retry-max-time 60 \
    "https://api.github.com/repos/darktable-org/darktable/releases/latest" | jq -r '.tag_name')
fi

set -euo pipefail

: "${DEB_HOST_ARCH:=$(dpkg --print-architecture)}"
: "${DEB_HOST_GNU_TYPE:=$(dpkg-architecture -a "$DEB_HOST_ARCH" -qDEB_HOST_GNU_TYPE)}"
CACHE_DIR="${BUILD_CACHE_DIR:-/build-cache}/Darktable-${DARKTABLE_VERSION}"
CACHE_MARKER="${CACHE_DIR}/Darktable-${DARKTABLE_VERSION}-complete"

# Check if this specific version is already built and cached
if [[ -f "$CACHE_MARKER" ]] && [[ -d "${CACHE_DIR}/output" ]]; then
  echo "Darktable ${DARKTABLE_VERSION} found in cache, reusing..."
  mkdir -p /output
  cp -ra "${CACHE_DIR}/output/"* /output/
  exit 0
fi

echo "Building Darktable ${DARKTABLE_VERSION} (cache miss)..."

echo Compiler: "${DEB_HOST_GNU_TYPE}" Arch: "${DEB_HOST_ARCH}"

apt-get install -y \
  clang \
  git \
  llvm \
  python3-jsonschema \
  libxml2-utils \
  intltool \
  iso-codes \
  xsltproc \
  libavif-dev:"${DEB_HOST_ARCH}" \
  libcairo2-dev:"${DEB_HOST_ARCH}" \
  libcolord-dev:"${DEB_HOST_ARCH}" \
  libcolord-gtk-dev:"${DEB_HOST_ARCH}" \
  libcurl4-gnutls-dev:"${DEB_HOST_ARCH}" \
  libexiv2-dev:"${DEB_HOST_ARCH}" \
  libgmic-dev:"${DEB_HOST_ARCH}" \
  libgphoto2-dev:"${DEB_HOST_ARCH}" \
  libgraphicsmagick1-dev:"${DEB_HOST_ARCH}" \
  libgtk-3-dev:"${DEB_HOST_ARCH}" \
  libjpeg-dev:"${DEB_HOST_ARCH}" \
  libjson-glib-dev:"${DEB_HOST_ARCH}" \
  libjxl-dev:"${DEB_HOST_ARCH}" \
  liblcms2-dev:"${DEB_HOST_ARCH}" \
  liblensfun-dev:"${DEB_HOST_ARCH}" \
  libopenexr-dev:"${DEB_HOST_ARCH}" \
  libopenjp2-7-dev:"${DEB_HOST_ARCH}" \
  libpng-dev:"${DEB_HOST_ARCH}" \
  libpugixml-dev:"${DEB_HOST_ARCH}" \
  librsvg2-dev:"${DEB_HOST_ARCH}" \
  libgtk-3-dev:"${DEB_HOST_ARCH}" \
  libsqlite3-dev:"${DEB_HOST_ARCH}" \
  libtiff-dev:"${DEB_HOST_ARCH}" \
  libwebp-dev:"${DEB_HOST_ARCH}"

URL="https://github.com/darktable-org/darktable.git"
echo download Darktable repo from "$URL"
git clone $URL darktable || true
cd darktable
git checkout ${DARKTABLE_VERSION}
git submodule init
git submodule update

./build.sh \
  --prefix "/opt/darktable" \
  --build-type "Release" \
  --install \
  --disable-kwallet \
  --disable-libsecret \
  --disable-lua \
  --disable-mac_integration \
  --disable-map \
  --disable-unity

exit 0

FEATURES="--with-heic --with-jpeg --with-png --with-raw --with-tiff --with-webp"

./configure \
  --enable-64bit-channel-masks \
  --enable-static --enable-shared --enable-delegate-build \
  --without-x --without-magick-plus-plus \
  --without-perl --disable-doc \
  --host="${DEB_HOST_GNU_TYPE}" \
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
cp -a /usr/local/etc/Darktable-7 /output/etc/
cp -a /usr/local/lib/Darktable-* /output/lib/
cp -a /usr/local/lib/libMagickCore-* /output/lib/
cp -a /usr/local/lib/libMagickWand-* /output/lib/
cp -a /usr/local/include/Darktable-7 /output/include/
cp -a /usr/local/lib/pkgconfig/Darktable*.pc /output/pkgconfig/
cp -a /usr/local/lib/pkgconfig/MagickCore*.pc /output/pkgconfig/
cp -a /usr/local/lib/pkgconfig/MagickWand*.pc /output/pkgconfig/
file /output/bin/magick

# After successful build, cache the results
echo "Caching Darktable ${DARKTABLE_VERSION} build results..."
mkdir -p "${CACHE_DIR}/output"
cp -ra /output/* "${CACHE_DIR}/output/"
touch "$CACHE_MARKER"

echo "Darktable ${DARKTABLE_VERSION} build complete and cached"
