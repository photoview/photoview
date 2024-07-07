#!/bin/sh

# Install Darktable if building for a supported architecture
if [ "${TARGETPLATFORM}" = "linux/amd64" ] || [ "${TARGETPLATFORM}" = "linux/arm64" ]; then \
  apt-get update
  apt-get install -y curl gnupg gpg

  echo 'deb https://download.opensuse.org/repositories/graphics:/darktable/Debian_12/ /' \
    | tee /etc/apt/sources.list.d/graphics:darktable.list
  curl -fsSL https://download.opensuse.org/repositories/graphics:/darktable/Debian_12/Release.key \
    | gpg --dearmor | tee /etc/apt/trusted.gpg.d/graphics_darktable.gpg > /dev/null

  apt-get update
  apt-get install -y darktable

  apt-get purge -y curl gnupg gpg
  apt-get autoremove -y
fi

apt-get update
apt-get install -y libdlib19.1 ffmpeg exiftool libheif1 sqlite3

# Remove build dependencies and cleanup
apt-get clean
rm -rf /var/lib/apt/lists/*
