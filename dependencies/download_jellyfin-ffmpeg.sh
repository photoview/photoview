#!/bin/bash
set -euo pipefail

: ${DEB_HOST_MULTIARCH=x86_64-linux-gnu}
: ${DEB_HOST_ARCH=amd64}

echo Compiler: ${DEB_HOST_MULTIARCH} Arch: ${DEB_HOST_ARCH}

curl -s https://api.github.com/repos/jellyfin/jellyfin-ffmpeg/releases/latest | grep \"browser_download_url\".*bookworm_${DEB_HOST_ARCH} | cut -d : -f 2,3 | tr -d ' ,"' | wget -i - -O ./jellyfin-ffmpeg.deb

mkdir -p /output/deb
cp ./jellyfin-ffmpeg.deb /output/deb/
