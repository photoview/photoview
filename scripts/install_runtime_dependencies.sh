#!/bin/bash

BUILD_DEPENDS=(gpg)

apt update
apt install -y ${BUILD_DEPENDS[@]} curl libdlib19.1 ffmpeg exiftool libheif1

# Install Darktable if building for a supported architecture
if [ "${TARGETPLATFORM}" = "linux/amd64" ] || [ "${TARGETPLATFORM}" = "linux/arm64" ]; then
  echo 'deb https://download.opensuse.org/repositories/graphics:/darktable/Debian_12/ /' > /etc/apt/sources.list.d/darktable.list
  curl -fsSL https://download.opensuse.org/repositories/graphics:/darktable/Debian_12/Release.key \
    | gpg --dearmor -o /etc/apt/trusted.gpg.d/darktable.gpg
  gpg --show-keys --with-fingerprint --dry-run /etc/apt/trusted.gpg.d/darktable.gpg

  apt update
  apt install -y darktable

  rm /etc/apt/sources.list.d/darktable.list /etc/apt/trusted.gpg.d/darktable.gpg
fi

# Remove build dependencies and cleanup
apt purge -y ${BUILD_DEPENDS[@]}
apt autoremove -y
apt clean
rm -rf /var/lib/apt/lists/*
