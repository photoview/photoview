# Build UI
FROM --platform=${BUILDPLATFORM:-linux/amd64} node:10 as ui

ARG API_ENDPOINT
ENV API_ENDPOINT=${API_ENDPOINT}

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

# Build API
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.14-alpine AS api

# Convert TARGETPLATFORM to GOARCH format
# https://github.com/tonistiigi/xx
COPY --from=tonistiigi/xx:golang / /
ARG TARGETPLATFORM

RUN mkdir -p /app
WORKDIR /app

# Download dependencies
COPY api/go.mod api/go.sum /app/
RUN go mod download

# Copy api source
COPY api /app

RUN go env && go build -v -o photoview .

# Copy api and ui to production environment
FROM alpine:3.12

# Install darktable for converting RAW images
RUN apk --no-cache add darktable

# Install ffmpeg for encoding videos
RUN apk --no-cache add ffmpeg

COPY --from=ui /app/dist /ui
COPY --from=api /app/database/migrations /database/migrations
COPY --from=api /app/photoview /app/photoview

ENV API_LISTEN_IP 127.0.0.1
ENV API_LISTEN_PORT 80

ENV SERVE_UI 1

EXPOSE 80

ENTRYPOINT ["/app/photoview"]
