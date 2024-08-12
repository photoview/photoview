#!/bin/bash

apt-get update
apt-get -t testing install -y imagemagick
apt-get install -y curl libdlib19.1 ffmpeg exiftool libheif1

convert -list format
convert -version

# Remove build dependencies and cleanup
apt-get purge -y ${BUILD_DEPENDS[@]}
apt-get autoremove -y
apt-get clean
rm -rf /var/lib/apt/lists/*
