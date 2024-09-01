#!/bin/sh

apk update

apk add curl ffmpeg exiftool libheif imagemagick lapack cblas
apk add dlib --repository=https://dl-cdn.alpinelinux.org/alpine/edge/testing

# Remove build dependencies and cleanup
apk cache clean
rm -rf /var/cache/apk/*
