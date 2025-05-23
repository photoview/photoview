#!/bin/bash
set -euo pipefail

: ${DEB_HOST_MULTIARCH=`uname -m`-linux-gnu}
: ${DEB_HOST_ARCH=`dpkg --print-architecture`}

echo Compiler: ${DEB_HOST_MULTIARCH} Arch: ${DEB_HOST_ARCH}

apt-get install -y libjxl-dev:${DEB_HOST_ARCH} liblcms2-dev:${DEB_HOST_ARCH} liblqr-1-0-dev:${DEB_HOST_ARCH} libdjvulibre-dev:${DEB_HOST_ARCH} libjpeg62-turbo-dev:${DEB_HOST_ARCH} libopenjp2-7-dev:${DEB_HOST_ARCH} libopenexr-dev:${DEB_HOST_ARCH} libpng-dev:${DEB_HOST_ARCH} libtiff-dev:${DEB_HOST_ARCH} libwebp-dev:${DEB_HOST_ARCH} libxml2-dev:${DEB_HOST_ARCH} libfftw3-dev:${DEB_HOST_ARCH} zlib1g-dev:${DEB_HOST_ARCH} liblzma-dev:${DEB_HOST_ARCH} libbz2-dev:${DEB_HOST_ARCH}
URL=$(curl -s https://api.github.com/repos/ImageMagick/ImageMagick/releases/latest | grep "tarball_url" | cut -d : -f 2,3 | tr -d ' ,"')
echo download ImageMagick from $URL
curl -L -o ./magick.tar.gz "$URL"

tar xfv ./magick.tar.gz
cd ImageMagick-*
./configure \
  --enable-64bit-channel-masks \
  --enable-static --enable-shared --enable-delegate-build \
  --with-x=no --with-magick-plus-plus=no --with-gvc=no \
  --without-perl --disable-doc \
  --host=${DEB_HOST_MULTIARCH}
make
make install
cd ..

mkdir -p /output/bin /output/etc /output/lib /output/include /output/pkgconfig
cp /usr/local/bin/magick /output/bin/
cp -a /usr/local/etc/ImageMagick-7 /output/etc/
cp -a /usr/local/lib/ImageMagick-* /output/lib/
cp -a /usr/local/lib/libMagickCore-* /output/lib/
cp -a /usr/local/lib/libMagickWand-* /output/lib/
cp -a /usr/local/include/ImageMagick-7 /output/include/
cp -a /usr/local/lib/pkgconfig/ImageMagick*.pc /output/pkgconfig/
cp -a /usr/local/lib/pkgconfig/MagickCore*.pc /output/pkgconfig/
cp -a /usr/local/lib/pkgconfig/MagickWand*.pc /output/pkgconfig/
file /output/bin/magick
