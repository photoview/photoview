#!/bin/bash
set -e

if [ "$TARGETPLATFORM" == "linux/arm64" ]; then
  dpkg --add-architecture arm64
  DEBIAN_ARCH='arm64'
elif [ "$TARGETPLATFORM" == "linux/arm/v6" ] || [ "$TARGETPLATFORM" == "linux/arm/v7" ]; then
  dpkg --add-architecture armhf
  DEBIAN_ARCH='armhf'
else
  dpkg --add-architecture amd64
  DEBIAN_ARCH='amd64'
fi

BUILD_DEPENDS=(gnupg2 gpg)

apt update
apt install -y ${BUILD_DEPENDS[@]} curl libdlib19.1:${DEBIAN_ARCH} exiftool:${DEBIAN_ARCH} libheif1:${DEBIAN_ARCH}

# Install Darktable if building for a supported architecture
# linux/arm64 can't install `darktable` with debian stable because it lacks packages:
#  > [linux/amd64->arm64 final 3/7] RUN groupadd -g 999 photoview   && useradd -r -u 999 -g photoview -m photoview   && chmod +x /app/scripts/*.sh   && /app/scripts/install_runtime_dependencies.sh:
# 64.89 Reading state information...
# 65.10 Some packages could not be installed. This may mean that you have
# 65.10 requested an impossible situation or if you are using the unstable
# 65.10 distribution that some required packages have not yet been created
# 65.10 or been moved out of Incoming.
# 65.10 The following information may help to resolve the situation:
# 65.10 
# 65.10 The following packages have unmet dependencies:
# 65.27  darktable:arm64 : Depends: libjs-scriptaculous:arm64 but it is not installable
# 65.28 E: Unable to correct problems, you have held broken packages.
if [ "${TARGETPLATFORM}" = "linux/amd64" ] ; then
  echo 'deb https://download.opensuse.org/repositories/graphics:/darktable/Debian_12/ /' \
    | tee /etc/apt/sources.list.d/graphics:darktable.list
  curl -fsSL https://download.opensuse.org/repositories/graphics:/darktable/Debian_12/Release.key \
    | gpg --dearmor | tee /etc/apt/trusted.gpg.d/graphics_darktable.gpg > /dev/null

  apt-get update
  apt-get install -y darktable:${DEBIAN_ARCH}
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
  apt install -y ffmpeg:${DEBIAN_ARCH}
fi

# Remove build dependencies and cleanup
apt-get purge -y ${BUILD_DEPENDS[@]}
apt-get autoremove -y
apt-get clean
rm -rf /var/lib/apt/lists/*
