#!/bin/sh
set -eu

apt-get update
apt-get install -y --no-install-recommends curl file libimage-exiftool-perl

# libheif dependencies
apt-get install -y --no-install-recommends libdav1d7 librav1e0.7 libde265-0 libx265-215 libjpeg62-turbo libopenh264-8 libpng16-16t64 libnuma1 zlib1g

# libraw dependencies
apt-get install -y --no-install-recommends libjpeg62-turbo liblcms2-2 zlib1g libgomp1

# ImageMagick dependencies
apt-get install -y --no-install-recommends libjxl0.11 liblcms2-2 liblqr-1-0 libdjvulibre21 libjpeg62-turbo libopenjp2-7 libopenexr-3-1-30 libpng16-16t64 libtiff6 libwebpmux3 libwebpdemux2 libwebp7 libxml2 zlib1g liblzma5 libbz2-1.0 libgomp1

# go-face dependencies
apt-get install -y --no-install-recommends libdlib19.2 libblas3 liblapack3 libjpeg62-turbo

# gomagic dependencies
apt-get install -y --no-install-recommends libmagic1t64
