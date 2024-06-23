### Build UI ###
FROM --platform=${BUILDPLATFORM:-linux/amd64} node:18 as ui

ARG REACT_APP_API_ENDPOINT
ENV REACT_APP_API_ENDPOINT=${REACT_APP_API_ENDPOINT}

# Set environment variable UI_PUBLIC_URL from build args, uses "/" as default
ARG UI_PUBLIC_URL
ENV UI_PUBLIC_URL=${UI_PUBLIC_URL:-/}

ARG VERSION
ENV VERSION=${VERSION:-undefined}
ENV REACT_APP_BUILD_VERSION=${VERSION:-undefined}

ARG BUILD_DATE
ENV BUILD_DATE=${BUILD_DATE:-undefined}
ENV REACT_APP_BUILD_DATE=${BUILD_DATE:-undefined}

ARG COMMIT_SHA
ENV COMMIT_SHA=${COMMIT_SHA:-}
ENV REACT_APP_BUILD_COMMIT_SHA=${COMMIT_SHA:-}

# Download dependencies
COPY ui /app
WORKDIR /app
RUN npm ci --omit=dev --ignore-scripts \
  # Build frontend
  && npm run build -- --base=$UI_PUBLIC_URL

### Build API ###
FROM --platform=${BUILDPLATFORM:-linux/amd64} debian:bookworm AS api
ARG TARGETPLATFORM

COPY docker/install_build_dependencies.sh /tmp/
COPY docker/go_wrapper.sh /go/bin/go
COPY api /app
WORKDIR /app

ENV GOPATH="/go"
ENV PATH="${GOPATH}/bin:${PATH}"
ENV CGO_ENABLED 1

# Download dependencies
RUN chmod +x /tmp/install_build_dependencies.sh \
  && chmod +x /go/bin/go \
  && /tmp/install_build_dependencies.sh \
  && go env \
  && go mod download \
  # Patch go-face
  && sed -i 's/-march=native//g' ${GOPATH}/pkg/mod/github.com/!kagami/go-face*/face.go \
  # Build dependencies that use CGO
  && go install \
    github.com/mattn/go-sqlite3 \
    github.com/Kagami/go-face \
  # Build api source
  && go build -v -o photoview .

### Copy api and ui to production environment ###
FROM debian:bookworm-slim
ARG TARGETPLATFORM

# See for details: https://github.com/hadolint/hadolint/wiki/DL4006
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
# Create a user to run Photoview server
RUN groupadd -g 999 photoview \
  && useradd -r -u 999 -g photoview -m photoview \
  # Required dependencies
  && apt-get update \
  && apt-get install -y curl gnupg gpg libdlib19.1 ffmpeg exiftool libheif1 sqlite3 \
  # Install Darktable if building for a supported architecture
  && if [ "${TARGETPLATFORM}" = "linux/amd64" ] || [ "${TARGETPLATFORM}" = "linux/arm64" ]; then \
    echo 'deb https://download.opensuse.org/repositories/graphics:/darktable/Debian_12/ /' \
      | tee /etc/apt/sources.list.d/graphics:darktable.list; \
    curl -fsSL https://download.opensuse.org/repositories/graphics:/darktable/Debian_12/Release.key \
      | gpg --dearmor | tee /etc/apt/trusted.gpg.d/graphics_darktable.gpg > /dev/null; \
    apt-get update; \
    apt-get install -y darktable; \
  fi \
  # Remove build dependencies and cleanup
  && apt-get purge -y gnupg gpg \
  && apt-get autoremove -y \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /home/photoview
COPY api/data /app/data
COPY --from=ui /app/dist /app/ui
COPY --from=api /app/photoview /app/photoview

ENV PHOTOVIEW_LISTEN_IP 127.0.0.1
ENV PHOTOVIEW_LISTEN_PORT 80

ENV PHOTOVIEW_SERVE_UI 1
ENV PHOTOVIEW_UI_PATH /app/ui
ENV PHOTOVIEW_FACE_RECOGNITION_MODELS_PATH /app/data/models
ENV PHOTOVIEW_MEDIA_CACHE /home/photoview/media-cache

EXPOSE ${PHOTOVIEW_LISTEN_PORT}

HEALTHCHECK --interval=60s --timeout=10s \
  CMD curl --fail http://localhost:${PHOTOVIEW_LISTEN_PORT}/api/graphql \
    -X POST -H 'Content-Type: application/json' \
    --data-raw '{"operationName":"CheckInitialSetup","variables":{},"query":"query CheckInitialSetup { siteInfo { initialSetup }}"}' \
    || exit 1

USER photoview
ENTRYPOINT ["/app/photoview"]
