#!/bin/sh
set -eu

apt-get update
apt-get install -y --no-install-recommends curl libimage-exiftool-perl

# libheif dependencies
apt-get install -y --no-install-recommends libdav1d7 librav1e0.7 libde265-0 libx265-215 libjpeg62-turbo libopenjp2-7 libopenh264-8 libpng16-16t64 libnuma1 zlib1g

# libraw dependencies
apt-get install -y --no-install-recommends libjpeg62-turbo liblcms2-2 zlib1g libgomp1

# ImageMagick dependencies
apt-get install -y --no-install-recommends libjxl0.11 liblcms2-2 liblqr-1-0 libdjvulibre21 libjpeg62-turbo libopenjp2-7 libopenexr-3-1-30 libpng16-16t64 libtiff6 libwebpmux3 libwebpdemux2 libwebp7 libxml2 zlib1g liblzma5 libbz2-1.0 libgomp1

# Darktable dependencies
apt-get install -y --no-install-recommends libavif16 libcurl4t64 libexiv2-28 libgmic1 libgraphicsmagick-q16-3t64 libgtk-3-0t64 libicu76 libjpeg62-turbo libjson-glib-1.0-0 libjxl0.11 liblcms2-2 liblensfun1 libopenexr-3-1-30 libopenjp2-7 libpng16-16t64 libpugixml1v5 librsvg2-2 libsqlite3-0 libtiff6 libwebp7

# go-face dependencies
apt-get install -y --no-install-recommends libdlib19.2 libblas3 liblapack3 libjpeg62-turbo
