#!/bin/sh
set -eu

: ${DEB_HOST_MULTIARCH=`uname -m`-linux-gnu}
: ${DEB_HOST_ARCH=`dpkg --print-architecture`}

echo Compiler: ${DEB_HOST_MULTIARCH} Arch: ${DEB_HOST_ARCH}

apt-get install -y zlib1g-dev:${DEB_HOST_ARCH} liblcms2-dev:${DEB_HOST_ARCH} libjpeg-dev:${DEB_HOST_ARCH}
curl -s https://api.github.com/repos/LibRaw/LibRaw/releases/latest | grep "tarball_url" | cut -d : -f 2,3 | tr -d ' ,"' | wget -i - -O ./libraw.tar.gz
tar xfv ./libraw.tar.gz
cd LibRaw-*
autoreconf --install
./configure --disable-option-checking --disable-silent-rules --disable-maintainer-mode --disable-dependency-tracking --host=${DEB_HOST_MULTIARCH}
make
make install
cd ..

mkdir -p /output/lib /output/include
cp -a /usr/local/lib/libraw_r* /output/lib/
cp -a /usr/local/include/libraw /output/include/
file /usr/local/lib/libraw_r.so*
