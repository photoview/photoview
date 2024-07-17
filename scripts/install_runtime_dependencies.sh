#!/bin/bash
set -e

BUILD_DEPENDS=(gnupg2 gpg)

apt update
apt install -y ${BUILD_DEPENDS[@]} curl libdlib19.1 exiftool libheif1

# Install Darktable if building for a supported architecture
if [ "${TARGETPLATFORM}" = "linux/amd64" ] || [ "${TARGETPLATFORM}" = "linux/arm64" ]; then
  echo 'deb https://download.opensuse.org/repositories/graphics:/darktable/Debian_12/ /' \
    | tee /etc/apt/sources.list.d/graphics:darktable.list
  curl -fsSL https://download.opensuse.org/repositories/graphics:/darktable/Debian_12/Release.key \
    | gpg --dearmor | tee /etc/apt/trusted.gpg.d/graphics_darktable.gpg > /dev/null

  apt-get update
  apt-get install -y darktable
fi

if [ ! -z "$TARGETPLATFORM" ]; then
  TARGETOS="$(echo $TARGETPLATFORM | cut -d"/" -f1)"
  TARGETARCH="$(echo $TARGETPLATFORM | cut -d"/" -f2)"
  TARGETVARIANT="$(echo $TARGETPLATFORM | cut -d"/" -f3)"
fi

if [ "$TARGETARCH" = "arm" ]; then
  TARGETARCH="armhf"
fi

JELLYFIN_FFMPEG_URL=$(curl https://api.github.com/repos/jellyfin/jellyfin-ffmpeg/releases/latest -s | grep "browser_download_url.*jellyfin-ffmpeg.*-bookworm_${TARGETARCH}\.deb" | cut -d '"' -f 4)
if [ "$JELLYFIN_FFMPEG_URL" != "" ]; then
  echo Install jellyfin-ffmpeg from \"${JELLYFIN_FFMPEG_URL}\" for arch \"${TARGETARCH}\"
  curl -L -o /tmp/jellyfin-ffmpeg.deb "${JELLYFIN_FFMPEG_URL}"
  apt install -y /tmp/jellyfin-ffmpeg.deb
  rm /tmp/jellyfin-ffmpeg.deb
  ln -s /usr/lib/jellyfin-ffmpeg/ffmpeg /usr/bin/ffmpeg
else
  echo Install ffmpeg from Debian repo for arch \"${TARGETARCH|}\"
  apt install -y ffmpeg
fi

# Remove build dependencies and cleanup
apt-get purge -y ${BUILD_DEPENDS[@]}
apt-get autoremove -y
apt-get clean
rm -rf /var/lib/apt/lists/*
