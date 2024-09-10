FROM --platform=${BUILDPLATFORM:-linux/amd64} debian:bookworm-slim AS build
ARG TARGETPLATFORM

# See for details: https://github.com/hadolint/hadolint/wiki/DL4006
SHELL ["/bin/bash", "-euo", "pipefail", "-c"]

RUN mkdir -p /tmp/build
WORKDIR /tmp/build

COPY prepare.sh /tmp/build/
RUN ./prepare.sh

COPY build_libraw.sh /tmp/build/
RUN export $(cat /env) && ./build_libraw.sh
COPY build_libheif.sh /tmp/build/
RUN export $(cat /env) && ./build_libheif.sh
COPY build_imagemagick.sh /tmp/build/
RUN export $(cat /env) && ./build_imagemagick.sh
COPY download_jellyfin-ffmpeg.sh /tmp/build/
RUN export $(cat /env) && ./download_jellyfin-ffmpeg.sh

COPY output.sh /
RUN /output.sh

FROM scratch AS release
COPY --from=build /artifacts.tar.gz /
