#!/bin/sh
set -eu

: ${DEB_HOST_MULTIARCH=`uname -m`-linux-gnu}
: ${DEB_HOST_ARCH=`dpkg --print-architecture`}

echo Compiler: ${DEB_HOST_MULTIARCH} Arch: ${DEB_HOST_ARCH}

apt-get install -y libaom-dev:${DEB_HOST_ARCH} libavcodec-dev:${DEB_HOST_ARCH} libdav1d-dev:${DEB_HOST_ARCH} libde265-dev:${DEB_HOST_ARCH} libgdk-pixbuf-2.0-dev:${DEB_HOST_ARCH} libjpeg-dev:${DEB_HOST_ARCH} libopenjp2-7-dev:${DEB_HOST_ARCH} libpng-dev:${DEB_HOST_ARCH} librav1e-dev:${DEB_HOST_ARCH} libsvtav1enc-dev:${DEB_HOST_ARCH} libx265-dev:${DEB_HOST_ARCH}
curl -s https://api.github.com/repos/strukturag/libheif/releases/latest | grep "tarball_url" | cut -d : -f 2,3 | tr -d ' ,"' | wget -i - -O ./libheif.tar.gz
tar xfv ./libheif.tar.gz
cd *-libheif-*
cmake --preset=release -DCMAKE_SYSTEM_PROCESSOR=${DEB_HOST_ARCH} -DCMAKE_C_COMPILER=${DEB_HOST_MULTIARCH}-gcc -DCMAKE_CXX_COMPILER=${DEB_HOST_MULTIARCH}-g++ -DPKG_CONFIG_EXECUTABLE=${DEB_HOST_MULTIARCH}-pkg-config -DCMAKE_LIBRARY_ARCHITECTURE=${DEB_HOST_MULTIARCH} -DPLUGIN_DIRECTORY=/usr/lib/${DEB_HOST_MULTIARCH}/libheif/plugins .
make
make install
cd ..

mkdir -p /output/lib /output/include
cp -a /usr/local/lib/libheif* /output/lib/
cp -a /usr/local/include/libheif /output/include/
file /usr/local/lib/libheif.so*
