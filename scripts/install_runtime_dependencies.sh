#!/bin/sh
set -eu

apt-get update
apt-get install -y curl exiftool

# libheif dependencies: no dependency
apt-get install -y libaom3 libdav1d6 libde265-0 libjpeg62-turbo libpng16-16 libx265-199
# libraw dependencies
apt-get install -y libjpeg62-turbo liblcms2-2
# ImageMagick dependencies
apt-get install -y libjxl0.7 liblcms2-2 liblqr-1-0 libdjvulibre21 libjpeg62-turbo libopenjp2-7 libopenexr-3-1-30 libpng16-16 libtiff6 libwebpmux3 libwebpdemux2 libwebp7 libxml2 libfreetype6 libgomp1
# go-face dependencies
apt-get install -y libdlib19.1 libblas3 liblapack3 libjpeg62-turbo
