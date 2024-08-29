#!/bin/bash
set -euo pipefail

apt-get update
apt-get install -y curl libdlib19.2 exiftool libheif1 imagemagick xz-utils

if [ "$(dpkg --print-architecture)" = "amd64" ]; then
  JELLYFIN_FFMPEG_URL=$(curl https://api.github.com/repos/jellyfin/jellyfin-ffmpeg/releases/latest -s | grep "browser_download_url.*jellyfin-ffmpeg_.*_portable_linux64-gpl.tar.xz" | cut -d '"' -f 4)
  if [ "${JELLYFIN_FFMPEG_URL}" = "" ]; then
    echo "Can't find jellyfin-ffmpeg download url"
    exit -1
  fi

  echo Install jellyfin-ffmpeg from \"${JELLYFIN_FFMPEG_URL}\"
  curl -L -o /tmp/jellyfin-ffmpeg.tar.xz "${JELLYFIN_FFMPEG_URL}"
  tar xfv /tmp/jellyfin-ffmpeg.tar.xz
  rm /tmp/jellyfin-ffmpeg.tar.xz
  mv ffmpeg ffprobe /usr/bin/
else
  echo Install ffmpeg from Debian repo
  apt install -y ffmpeg
fi

# Remove build dependencies and cleanup
apt-get autoremove -y
apt-get clean
rm -rf /var/lib/apt/lists/*
