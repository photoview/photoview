#!/bin/sh
set -eu

: ${DEB_HOST_ARCH=`dpkg --print-architecture`}
echo Target Arch: ${DEB_HOST_ARCH}
echo Env Arch: $(dpkg --print-architecture)

# Download Darktable
if [ ! -f "/output/deb/darktable.deb" ]
then
  DARKTABLE_URL="http://download.opensuse.org/repositories/graphics:/darktable/Debian_13"
  echo "deb ${DARKTABLE_URL}/ /" | tee /etc/apt/sources.list.d/graphics:darktable.list
  curl -fsSL "${DARKTABLE_URL}/Release.key" | gpg --dearmor | tee /etc/apt/trusted.gpg.d/graphics_darktable.gpg > /dev/null
  apt-get update
  echo download darktable from "${DARKTABLE_URL}"
  apt-get download darktable:${DEB_HOST_ARCH}

  mkdir -p /output/deb
  cp darktable*.deb /output/deb/darktable.deb
fi

# Download FFMpeg
if [ ! -f "/output/deb/jellyfin-ffmpeg.deb" ]
then
  JELLYFIN_FFMPEG_VERSION=$(curl -fsSL --retry 2 --retry-delay 5 --retry-max-time 60 \
    "https://api.github.com/repos/jellyfin/jellyfin-ffmpeg/releases/latest" | jq -r '.tag_name')

  VER="${JELLYFIN_FFMPEG_VERSION#v}"
  MAJOR_VER=$(echo "${VER}" | cut -d. -f1)
  FFMPEG_URL="https://github.com/jellyfin/jellyfin-ffmpeg/releases/download/${JELLYFIN_FFMPEG_VERSION}/jellyfin-ffmpeg${MAJOR_VER}_${VER}-trixie_${DEB_HOST_ARCH}.deb"
  apt-get install -y --no-install-recommends curl ca-certificates
  echo download jellyfin-ffmpeg from "${FFMPEG_URL}"
  mkdir -p /output/deb
  curl -fsSL --retry 2 --retry-delay 5 --retry-max-time 60 -o /output/deb/jellyfin-ffmpeg.deb "${FFMPEG_URL}"
fi

# Don't install in cross-build environment because it can't run.
# Since we won't develop in cross-build environment, it's fine of just downloading the deb file.
# Downloaded deb file will be installed in the release stage, which runs in an arch-native environment.
if [ "${DEB_HOST_ARCH}" != "$(dpkg --print-architecture)" ]
then
  exit 0
fi

# Install binary dependencies for test in the native environment
apt-get update

# exiftool
apt-get install -y --no-install-recommends libimage-exiftool-perl:${DEB_HOST_ARCH}

# graphicswand
apt-get install -y --no-install-recommends libgraphicsmagick-q16-3t64:${DEB_HOST_ARCH}

# go-face dependencies
apt-get install -y --no-install-recommends libdlib19.2:${DEB_HOST_ARCH} libblas3:${DEB_HOST_ARCH} liblapack3:${DEB_HOST_ARCH} libjpeg62-turbo:${DEB_HOST_ARCH}

# darktable
apt-get install -y --no-install-recommends /output/deb/darktable.deb

# ffmpeg
apt-get install -y --no-install-recommends /output/deb/jellyfin-ffmpeg.deb
ln -s /usr/lib/jellyfin-ffmpeg/ffmpeg /usr/local/bin/
ln -s /usr/lib/jellyfin-ffmpeg/ffprobe /usr/local/bin/
