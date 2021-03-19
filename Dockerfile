### Build UI ###
FROM --platform=${BUILDPLATFORM:-linux/amd64} node:10 as ui

ARG PHOTOVIEW_API_ENDPOINT
ENV PHOTOVIEW_API_ENDPOINT=${PHOTOVIEW_API_ENDPOINT}

# Set environment variable UI_PUBLIC_URL from build args, uses "/" as default
ARG UI_PUBLIC_URL
ENV UI_PUBLIC_URL=${UI_PUBLIC_URL:-/}

RUN mkdir -p /app
WORKDIR /app

# Download dependencies
COPY ui/package*.json /app/
RUN npm install

# Build frontend
COPY ui /app
RUN npm run build -- --public-url $UI_PUBLIC_URL

### Build API ###
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.15-buster AS api

# Install G++/GCC cross compilers
RUN dpkg --add-architecture arm64 && dpkg --add-architecture armhf
RUN apt-get update && apt-get install -y \
  g++-aarch64-linux-gnu \
  libc6-dev-arm64-cross \
  g++-arm-linux-gnueabihf \
  libc6-dev-armhf-cross \
  # Install go-face dependencies
  libdlib-dev libdlib-dev:arm64 libdlib-dev:armhf \
  libblas-dev libblas-dev:arm64 libblas-dev:armhf \
  liblapack-dev liblapack-dev:arm64 liblapack-dev:armhf \
  libjpeg62-turbo-dev libjpeg62-turbo-dev:arm64 libjpeg62-turbo-dev:armhf \
  # Cleanup
  && apt-get clean && rm -rf /var/lib/apt/lists/*

COPY docker/go_wrapper.sh /go/bin/go
RUN chmod +x /go/bin/go

ENV CGO_ENABLED 1

ARG TARGETPLATFORM
RUN go env

RUN mkdir -p /app
WORKDIR /app

# Download dependencies
COPY api/go.mod api/go.sum /app/
RUN go mod download

# Patch go-face
RUN sed -i 's/-march=native//g' /go/pkg/mod/github.com/!kagami/go-face*/face.go

# Build dependencies that use CGO
RUN go install \
  github.com/mattn/go-sqlite3 \
  github.com/Kagami/go-face

# Copy and build api source
COPY api /app
RUN go build -v -o photoview .

### Copy api and ui to production environment ###
FROM debian:buster
ARG TARGETPLATFORM
WORKDIR /app

COPY api/data /app/data

RUN apt-get update \
  # Required dependencies
  && apt-get install -y curl gpg libdlib19 ffmpeg exiftool

# Install Darktable if building for a supported architecture
RUN if [ "${TARGETPLATFORM}" = "linux/amd64" ] || [ "${TARGETPLATFORM}" = "linux/arm64" ]; then \
  echo 'deb http://download.opensuse.org/repositories/graphics:/darktable/Debian_10/ /' | tee /etc/apt/sources.list.d/graphics:darktable.list \
  && curl -fsSL https://download.opensuse.org/repositories/graphics:darktable/Debian_10/Release.key | gpg --dearmor | tee /etc/apt/trusted.gpg.d/graphics_darktable.gpg > /dev/null \
  && apt-get update \
  && apt-get install -y darktable; \
  fi

# Remove build dependencies and cleanup
RUN apt-get purge -y curl gpg \
  && apt-get autoremove -y \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

COPY --from=ui /app/dist /ui
COPY --from=api /app/photoview /app/photoview

ENV PHOTOVIEW_LISTEN_IP 127.0.0.1
ENV PHOTOVIEW_LISTEN_PORT 80

ENV PHOTOVIEW_SERVE_UI 1
ENV PHOTOVIEW_UI_PATH /ui

EXPOSE 80

ENTRYPOINT ["/app/photoview"]
