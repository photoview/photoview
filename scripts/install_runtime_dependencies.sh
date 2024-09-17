#!/bin/sh
set -eu

apt-get update
apt-get install -y curl exiftool

# libheif dependencies: no dependency
# libraw dependencies
apt-get install -y libjpeg62-turbo liblcms2-2
# ImageMagick dependencies
apt-get install -y libjbig0 libtiff6 libfreetype6 libjxl0.7 liblqr-1-0 libpng16-16 libdjvulibre21 libwebpmux3 libwebpdemux2 libwebp7 libopenexr-3-1-30 libopenjp2-7 libjpeg62-turbo liblcms2-2 libxml2 libx11-6 libgomp1
# go-face dependencies
apt-get install -y libdlib19.1 libblas3 liblapack3 libjpeg62-turbo
