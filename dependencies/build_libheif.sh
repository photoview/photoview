#!/bin/sh
set -eu

: ${DEB_HOST_MULTIARCH=`uname -m`-linux-gnu}
: ${DEB_HOST_ARCH=`dpkg --print-architecture`}

echo Compiler: ${DEB_HOST_MULTIARCH} Arch: ${DEB_HOST_ARCH}

apt-get install -y libaom-dev:${DEB_HOST_ARCH} libdav1d-dev:${DEB_HOST_ARCH} libde265-dev:${DEB_HOST_ARCH} libjpeg62-turbo-dev:${DEB_HOST_ARCH} libnuma-dev:${DEB_HOST_ARCH} libpng-dev:${DEB_HOST_ARCH} libx265-dev:${DEB_HOST_ARCH} zlib1g-dev:${DEB_HOST_ARCH}
URL=$(curl -s https://api.github.com/repos/strukturag/libheif/releases/latest | grep "tarball_url" | cut -d : -f 2,3 | tr -d ' ,"')
echo download libraw from $URL
curl -L -o ./libheif.tar.gz "$URL"

tar xfv ./libheif.tar.gz
cd *-libheif-*
cmake --preset=release -DCMAKE_SYSTEM_PROCESSOR=${DEB_HOST_ARCH} -DCMAKE_C_COMPILER=${DEB_HOST_MULTIARCH}-gcc -DCMAKE_CXX_COMPILER=${DEB_HOST_MULTIARCH}-g++ -DPKG_CONFIG_EXECUTABLE=${DEB_HOST_MULTIARCH}-pkg-config -DCMAKE_LIBRARY_ARCHITECTURE=${DEB_HOST_MULTIARCH} -DENABLE_PLUGIN_LOADING=OFF -DWITH_GDK_PIXBUF=OFF .
make
make install
cd ..

mkdir -p /output/bin /output/lib /output/include /output/pkgconfig
cp -a /usr/local/bin/heif-* /output/bin/
cp -a /usr/local/lib/libheif* /output/lib/
cp -a /usr/lib/${DEB_HOST_MULTIARCH}/libheif/plugins/libheif-*.so /output/lib
cp -a /usr/local/lib/${DEB_HOST_MULTIARCH}/gdk-pixbuf-*/*/loaders/libpixbufloader-heif.so /output/lib
cp -a /usr/local/lib/pkgconfig/libheif* /output/pkgconfig/
cp -a /usr/local/include/libheif /output/include/
file /usr/local/lib/libheif.so*
