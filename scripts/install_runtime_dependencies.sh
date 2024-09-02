#!/bin/sh
set -euo pipefail

apk update

apk add curl libheif lapack cblas exiftool
apk add ffmpeg ffmpeg-libs ffmpeg-libavcodec ffmpeg-libavformat
apk add imagemagick imagemagick-libs imagemagick-heic imagemagick-jpeg imagemagick-raw imagemagick-tiff imagemagick-webp imagemagick-svg imagemagick-jxl
apk add dlib --repository=https://dl-cdn.alpinelinux.org/alpine/edge/testing
