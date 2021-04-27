### Build UI ###
FROM --platform=${BUILDPLATFORM:-linux/amd64} node:15 as ui

ARG PHOTOVIEW_API_ENDPOINT
ENV PHOTOVIEW_API_ENDPOINT=${PHOTOVIEW_API_ENDPOINT}

# Set environment variable UI_PUBLIC_URL from build args, uses "/" as default
ARG UI_PUBLIC_URL
ENV UI_PUBLIC_URL=${UI_PUBLIC_URL:-/}

ARG VERSION
ENV VERSION=${VERSION:-undefined}

ARG BUILD_DATE
ENV BUILD_DATE=${BUILD_DATE:-undefined}

ARG COMMIT_SHA
ENV COMMIT_SHA=${COMMIT_SHA:-}

RUN mkdir -p /app
WORKDIR /app

# Download dependencies
COPY ui/package*.json /app/
RUN HUSKY=0 npm ci --only=production

# Build frontend
COPY ui /app
RUN npm run build -- --public-url $UI_PUBLIC_URL

### Build API ###
FROM alpine:edge AS api
# ARG TARGETPLATFORM

# COPY docker/install_build_dependencies.sh /tmp/
# RUN chmod +x /tmp/install_build_dependencies.sh && /tmp/install_build_dependencies.sh

# COPY docker/go_wrapper.sh /go/bin/go
# RUN chmod +x /go/bin/go
# ENV GOPATH="/go"
# ENV PATH="${GOPATH}/bin:${PATH}"

RUN apk add --no-cache \
  go openblas-dev lapack-dev jpeg-dev libheif-dev g++

RUN apk add --no-cache -X http://dl-cdn.alpinelinux.org/alpine/edge/testing \
  dlib

ENV CGO_ENABLED 1

# RUN go env

RUN mkdir -p /app
WORKDIR /app

# Download dependencies
COPY api/go.mod api/go.sum /app/
RUN go mod download

# Patch go-face
# RUN sed -i 's/-march=native//g' ${GOPATH}/pkg/mod/github.com/!kagami/go-face*/face.go
RUN sed -i 's/-lblas/-lopenblas/g' /root/go/pkg/mod/github.com/!kagami/go-face*/face.go

# Build dependencies that use CGO
RUN go install \
  github.com/mattn/go-sqlite3 \
  github.com/Kagami/go-face

# Copy and build api source
COPY api /app
RUN go build -v -o photoview .

### Copy api and ui to production environment ###
FROM alpine:edge
ARG TARGETPLATFORM
WORKDIR /app

COPY api/data /app/data

# RUN apt-get update \
#   # Required dependencies
#   && apt-get install -y curl gpg libdlib19 ffmpeg exiftool libheif1

RUN apk add --no-cache \
  ffmpeg exiftool libheif

RUN apk add --no-cache -X http://dl-cdn.alpinelinux.org/alpine/edge/testing \
  dlib

# Install Darktable if building for a supported architecture
# RUN if [ "${TARGETPLATFORM}" = "linux/amd64" ] || [ "${TARGETPLATFORM}" = "linux/arm64" ]; then \
#   apt-get install -y darktable; fi

# Remove build dependencies and cleanup
# RUN apt-get purge -y curl gpg \
#   && apt-get autoremove -y \
#   && apt-get clean \
#   && rm -rf /var/lib/apt/lists/*

COPY --from=ui /app/dist /ui
COPY --from=api /app/photoview /app/photoview

ENV PHOTOVIEW_LISTEN_IP 127.0.0.1
ENV PHOTOVIEW_LISTEN_PORT 80

ENV PHOTOVIEW_SERVE_UI 1
ENV PHOTOVIEW_UI_PATH /ui

EXPOSE 80

ENTRYPOINT ["/app/photoview"]
