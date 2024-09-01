#!/bin/sh

apk update

apk add curl libheif lapack cblas ffmpeg exiftool imagemagick imagemagick-heic imagemagick-jpeg imagemagick-raw imagemagick-tiff
apk add dlib --repository=https://dl-cdn.alpinelinux.org/alpine/edge/testing

# Remove build dependencies and cleanup
apk cache clean
rm -rf /var/cache/apk/*
