#!/bin/bash
set -euo pipefail

apt-get update
apt-get install -y curl libdlib19.2 exiftool libheif1 imagemagick

JELLYFIN_FFMPEG_URL=$(curl https://api.github.com/repos/jellyfin/jellyfin-ffmpeg/releases/latest -s | grep "browser_download_url.*jellyfin-ffmpeg.*-bookworm_${DEBIAN_ARCH}.deb" | cut -d '"' -f 4)
if [ "${JELLYFIN_FFMPEG_URL}" != "" ]; then
  echo Install jellyfin-ffmpeg from \"${JELLYFIN_FFMPEG_URL}\" for arch \"${DEBIAN_ARCH}\"
  curl -L -o /tmp/jellyfin-ffmpeg.deb "${JELLYFIN_FFMPEG_URL}"
  apt install -y /tmp/jellyfin-ffmpeg.deb
  rm /tmp/jellyfin-ffmpeg.deb
  ln -s /usr/lib/jellyfin-ffmpeg/ffmpeg /usr/bin/ffmpeg
  ln -s /usr/lib/jellyfin-ffmpeg/ffprobe /usr/bin/ffprobe
else
  echo Install ffmpeg from Debian repo for arch \"${DEBIAN_ARCH}\"
  apt install -y ffmpeg
fi

# Remove build dependencies and cleanup
apt-get autoremove -y
apt-get clean
rm -rf /var/lib/apt/lists/*
