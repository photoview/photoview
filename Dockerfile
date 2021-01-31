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
COPY ui /app

# Build frontend
RUN npm run build -- --public-url $UI_PUBLIC_URL

### Build API ###
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.15-buster AS api

# Install GCC cross compilers
RUN apt-get update
RUN apt-get install -y gcc-aarch64-linux-gnu libc6-dev-arm64-cross gcc-arm-linux-gnueabihf libc6-dev-armhf-cross

COPY --from=tonistiigi/xx:golang / /

ARG TARGETPLATFORM
RUN go env

RUN mkdir -p /app
WORKDIR /app

# Download dependencies
COPY api/go.mod api/go.sum /app/
RUN go mod download

# Build go-sqlite3 dependency with CGO
ENV CGO_ENABLED 1
RUN go install github.com/mattn/go-sqlite3

# Copy api source
COPY api /app

RUN go build -v -o photoview .

### Copy api and ui to production environment ###
FROM debian:buster

# Install darktable for converting RAW images, and ffmpeg for encoding videos
RUN apt-get update
RUN apt-get install -y darktable; exit 0
RUN apt-get install -y ffmpeg; exit 0
RUN rm -rf /var/lib/apt/lists/*

COPY --from=ui /app/dist /ui
COPY --from=api /app/photoview /app/photoview

ENV PHOTOVIEW_LISTEN_IP 127.0.0.1
ENV PHOTOVIEW_LISTEN_PORT 80

ENV PHOTOVIEW_SERVE_UI 1

EXPOSE 80

ENTRYPOINT ["/app/photoview"]
