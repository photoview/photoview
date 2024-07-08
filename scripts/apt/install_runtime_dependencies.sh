#!/bin/bash

BUILD_DEPENDS=(gnupg2 gpg)

apt-get update
apt-get install -y ${BUILD_DEPENDS[@]} curl libdlib19.1 ffmpeg exiftool libheif1

# Install Darktable if building for a supported architecture
if [ "${TARGETPLATFORM}" = "linux/amd64" ] || [ "${TARGETPLATFORM}" = "linux/arm64" ]; then
  echo 'deb https://download.opensuse.org/repositories/graphics:/darktable/Debian_12/ /' \
    | tee /etc/apt/sources.list.d/graphics:darktable.list
  curl -fsSL https://download.opensuse.org/repositories/graphics:/darktable/Debian_12/Release.key \
    | gpg --dearmor | tee /etc/apt/trusted.gpg.d/graphics_darktable.gpg > /dev/null

  apt-get update
  apt-get install -y darktable
fi

# Remove build dependencies and cleanup
apt-get purge -y ${BUILD_DEPENDS[@]}
apt-get autoremove -y
apt-get clean
rm -rf /var/lib/apt/lists/*
