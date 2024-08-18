#!/bin/bash
set -euo pipefail

apt-get update
apt-get install -y curl libdlib19.2 ffmpeg exiftool libheif1 imagemagick

# Remove build dependencies and cleanup
apt-get autoremove -y
apt-get clean
rm -rf /var/lib/apt/lists/*
