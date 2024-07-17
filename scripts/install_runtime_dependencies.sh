#!/bin/bash
set -e

DEBIAN_ARCH=$(dpkg --print-architecture)

BUILD_DEPENDS=(gnupg2 gpg)

apt update
apt install -y ${BUILD_DEPENDS[@]} curl libdlib19.1 exiftool libheif1

# Install Darktable if building for a supported architecture
if [ "${DEBIAN_ARCH}" = "amd64" ] || [ "${DEBIAN_ARCH}" = "arm64" ] ; then
  echo 'deb https://download.opensuse.org/repositories/graphics:/darktable/Debian_12/ /' \
    | tee /etc/apt/sources.list.d/graphics:darktable.list
  curl -fsSL https://download.opensuse.org/repositories/graphics:/darktable/Debian_12/Release.key \
    | gpg --dearmor | tee /etc/apt/trusted.gpg.d/graphics_darktable.gpg > /dev/null

  apt-get update
  apt-get install -y darktable
fi

JELLYFIN_FFMPEG_URL=$(curl https://api.github.com/repos/jellyfin/jellyfin-ffmpeg/releases/latest -s | grep "browser_download_url.*jellyfin-ffmpeg.*-bookworm_${DEBIAN_ARCH}.deb" | cut -d '"' -f 4)
if [ "${JELLYFIN_FFMPEG_URL}" != "" ]; then
  echo Install jellyfin-ffmpeg from \"${JELLYFIN_FFMPEG_URL}\" for arch \"${DEBIAN_ARCH}\"
  curl -L -o /tmp/jellyfin-ffmpeg.deb "${JELLYFIN_FFMPEG_URL}"
  apt install -y /tmp/jellyfin-ffmpeg.deb
  rm /tmp/jellyfin-ffmpeg.deb
  ln -s /usr/lib/jellyfin-ffmpeg/ffmpeg /usr/bin/ffmpeg
else
  echo Install ffmpeg from Debian repo for arch \"${DEBIAN_ARCH}\"
  apt install -y ffmpeg
fi

# Remove build dependencies and cleanup
apt-get purge -y ${BUILD_DEPENDS[@]}
apt-get autoremove -y
apt-get clean
rm -rf /var/lib/apt/lists/*
