#!/bin/sh
set -euo pipefail

apk update

apk add go g++ blas-dev cblas lapack-dev jpeg-dev libheif-dev
apk add dlib-dev --repository=https://dl-cdn.alpinelinux.org/alpine/edge/testing

# Install tools for development
apk add sqlite
go install github.com/cespare/reflex@latest
