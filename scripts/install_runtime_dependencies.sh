#!/bin/bash

apt-get update
apt-get install -y curl libdlib19.1 ffmpeg exiftool libheif1

# Install Darktable if building for a supported architecture
if [ "${TARGETPLATFORM}" = "linux/amd64" ] || [ "${TARGETPLATFORM}" = "linux/arm64" ]; then
  echo 'deb [trusted=true] https://download.opensuse.org/repositories/graphics:/darktable/Debian_12/ /' > /etc/apt/sources.list.d/darktable.list
  # Release key is invalid, just trust the repo
  curl -fsSL https://download.opensuse.org/repositories/graphics:/darktable/Debian_12/Release.key \
    | gpg --dearmor -o /etc/apt/trusted.gpg.d/darktable.gpg
  gpg --show-keys --with-fingerprint --dry-run /etc/apt/trusted.gpg.d/darktable.gpg

  apt-get update
  apt-get install -y darktable
fi

apt-get install -y libfontconfig libx11-6 libharfbuzz-bin libfribidi-bin
curl -o ./magick https://imagemagick.org/archive/binaries/magick
chmod +x ./magick
./magick --appimage-extract
cp -r ./squashfs-root/usr/* /usr/
rm -Rf ./squashfs-root ./magick

# Remove build dependencies and cleanup
apt-get purge -y ${BUILD_DEPENDS[@]}
apt-get autoremove -y
apt-get clean
rm -rf /var/lib/apt/lists/*
