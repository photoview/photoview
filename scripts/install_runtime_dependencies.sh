#!/bin/sh
set -eu

apt-get update
apt-get install -y curl exiftool

# libheif dependencies
apt-get install -y libdav1d6 librav1e0 libde265-0 libx265-199 libjpeg62-turbo libopenh264-7 libpng16-16 libnuma1 zlib1g

# libraw dependencies
apt-get install -y libjpeg62-turbo liblcms2-2 zlib1g libgomp1

# ImageMagick dependencies
apt-get install -y libjxl0.7 liblcms2-2 liblqr-1-0 libdjvulibre21 libjpeg62-turbo libopenjp2-7 libopenexr-3-1-30 libpng16-16 libtiff6 libwebpmux3 libwebpdemux2 libwebp7 libxml2 zlib1g liblzma5 libbz2-1.0 libgomp1

# go-face dependencies
apt-get install -y libdlib19.1 libblas3 liblapack3 libjpeg62-turbo

# gomagic dependencies
apt-get install -y libmagic1
