#!/bin/bash
set -euo pipefail

apt-get update
apt-get full-upgrade -y

apt-get -t testing install -y imagemagick curl libdlib19.2 ffmpeg exiftool libheif1

convert -version

# Remove build dependencies and cleanup
apt-get autoremove -y
apt-get clean
rm -rf /var/lib/apt/lists/*
