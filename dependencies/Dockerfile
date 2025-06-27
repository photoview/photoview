FROM --platform=${BUILDPLATFORM:-linux/amd64} debian:bookworm-slim AS base-prep
ARG TARGETPLATFORM
SHELL ["/bin/bash", "-euo", "pipefail", "-c"]
COPY --chmod=0755 prepare.sh /tmp/
RUN /tmp/prepare.sh \
    && mkdir -p /output/bin /output/lib /output/include /output/pkgconfig /output/etc /output/deb

FROM base-prep AS build-libraw
ARG LIBRAW_VERSION
ARG TARGETARCH
ENV BUILD_CACHE_DIR=/build-cache
WORKDIR /tmp/build-libraw
COPY --chmod=0755 build_libraw.sh .
RUN --mount=type=cache,target=${BUILD_CACHE_DIR},id=libraw-${TARGETARCH}-${LIBRAW_VERSION} \
    --mount=type=secret,id=github_token \
    if [ -f /run/secrets/github_token ]; then \
    export GITHUB_TOKEN=$(cat /run/secrets/github_token); \
    fi \
    && export $(cat /env) && ./build_libraw.sh

FROM base-prep AS build-libheif
ARG LIBHEIF_VERSION
ARG TARGETARCH
ENV BUILD_CACHE_DIR=/build-cache
WORKDIR /tmp/build-libheif
COPY --chmod=0755 build_libheif.sh .
RUN --mount=type=cache,target=${BUILD_CACHE_DIR},id=libheif-${TARGETARCH}-${LIBHEIF_VERSION} \
    --mount=type=secret,id=github_token \
    if [ -f /run/secrets/github_token ]; then \
    export GITHUB_TOKEN=$(cat /run/secrets/github_token); \
    fi \
    && export $(cat /env) && ./build_libheif.sh

FROM base-prep AS build-imagemagick
ARG IMAGEMAGICK_VERSION
ARG TARGETARCH
ENV BUILD_CACHE_DIR=/build-cache
WORKDIR /tmp/build-imagemagick
COPY --chmod=0755 build_imagemagick.sh .
RUN --mount=type=cache,target=${BUILD_CACHE_DIR},id=imagemagick-${TARGETARCH}-${IMAGEMAGICK_VERSION} \
    --mount=type=secret,id=github_token \
    if [ -f /run/secrets/github_token ]; then \
    export GITHUB_TOKEN=$(cat /run/secrets/github_token); \
    fi \
    && export $(cat /env) && ./build_imagemagick.sh

FROM base-prep AS download-jellyfin-ffmpeg
ARG JELLYFIN_FFMPEG_VERSION
ARG TARGETARCH
ENV BUILD_CACHE_DIR=/build-cache
WORKDIR /tmp/download-ffmpeg
COPY --chmod=0755 download_jellyfin-ffmpeg.sh .
RUN --mount=type=cache,target=${BUILD_CACHE_DIR},id=ffmpeg-${TARGETARCH}-${JELLYFIN_FFMPEG_VERSION} \
    export $(cat /env) && ./download_jellyfin-ffmpeg.sh

FROM base-prep AS final-assembly
COPY --from=build-libraw /output/ /output/
COPY --from=build-libheif /output/ /output/
COPY --from=build-imagemagick /output/ /output/
COPY --from=download-jellyfin-ffmpeg /output/ /output/
COPY --chmod=0755 output.sh /
RUN /output.sh

FROM scratch AS release
COPY --from=final-assembly /artifacts.tar.gz /
USER nobody
