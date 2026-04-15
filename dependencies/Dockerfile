FROM --platform=${BUILDPLATFORM:-linux/amd64} debian:trixie-slim AS base-prep
ARG TARGETPLATFORM
SHELL ["/bin/bash", "-euo", "pipefail", "-c"]
COPY --chmod=0755 prepare.sh /tmp/
RUN /tmp/prepare.sh \
    && mkdir -p /output/bin /output/lib /output/include /output/pkgconfig /output/etc /output/deb

FROM base-prep AS build-libraw
ARG LIBRAW_VERSION
ARG TARGETARCH
WORKDIR /tmp/build-libraw
COPY --chmod=0755 build_libraw.sh .
RUN --mount=type=secret,id=github_token \
    if [ -f /run/secrets/github_token ]; then \
      export GITHUB_TOKEN=$(cat /run/secrets/github_token); \
    fi \
    && set -a && source /env && set +a \
    && DEBIAN_FRONTEND=noninteractive ./build_libraw.sh

FROM build-libraw AS build-imagemagick
ARG IMAGEMAGICK_VERSION
ARG TARGETARCH
WORKDIR /tmp/build-imagemagick
COPY --chmod=0755 build_imagemagick.sh .
RUN --mount=type=secret,id=github_token \
    if [ -f /run/secrets/github_token ]; then \
      export GITHUB_TOKEN=$(cat /run/secrets/github_token); \
    fi \
    && set -a && source /env && set +a \
    && DEBIAN_FRONTEND=noninteractive ./build_imagemagick.sh

FROM base-prep AS download-jellyfin-ffmpeg
ARG JELLYFIN_FFMPEG_VERSION
ARG TARGETARCH
WORKDIR /tmp/download-ffmpeg
COPY --chmod=0755 download_jellyfin-ffmpeg.sh .
RUN set -a && source /env && set +a \
    && DEBIAN_FRONTEND=noninteractive ./download_jellyfin-ffmpeg.sh

FROM base-prep AS final-assembly
COPY --from=build-libraw /output/ /output/
COPY --from=build-imagemagick /output/ /output/
COPY --from=download-jellyfin-ffmpeg /output/ /output/
COPY --chmod=0755 output.sh /
RUN /output.sh

FROM debian:trixie-slim AS release
COPY --from=final-assembly /artifacts.tar.gz /
