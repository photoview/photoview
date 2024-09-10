#!/bin/sh
set -eu

: ${DEB_HOST_MULTIARCH=`uname -m`-linux-gnu}
: ${DEB_HOST_ARCH=`dpkg --print-architecture`}

echo Compiler: ${DEB_HOST_MULTIARCH} Arch: ${DEB_HOST_ARCH}

apt-get install -y libjxl-dev:${DEB_HOST_ARCH} libfftw3-dev:${DEB_HOST_ARCH} liblcms2-dev:${DEB_HOST_ARCH} liblqr-1-0-dev:${DEB_HOST_ARCH} zlib1g-dev:${DEB_HOST_ARCH} liblzma-dev:${DEB_HOST_ARCH} libbz2-dev:${DEB_HOST_ARCH} libdjvulibre-dev:${DEB_HOST_ARCH} libexif-dev:${DEB_HOST_ARCH} libjpeg-dev:${DEB_HOST_ARCH} libopenjp2-7-dev:${DEB_HOST_ARCH} libopenexr-dev:${DEB_HOST_ARCH} libpng-dev:${DEB_HOST_ARCH} libtiff-dev:${DEB_HOST_ARCH} libwmf-dev:${DEB_HOST_ARCH} libwebp-dev:${DEB_HOST_ARCH} libxml2-dev:${DEB_HOST_ARCH}
curl -s https://api.github.com/repos/ImageMagick/ImageMagick/releases/latest | grep "tarball_url" | cut -d : -f 2,3 | tr -d ' ,"' | wget -i - -O ./magick.tar.gz
tar xfv ./magick.tar.gz
cd ImageMagick-*
./configure --enable-static --enable-delegate-build --disable-shared --host=${DEB_HOST_MULTIARCH}
make
make install
cd ..

mkdir -p /output/bin
cp /usr/local/bin/magick /output/bin/
file /output/bin/magick
