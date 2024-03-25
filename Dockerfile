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

RUN mkdir -p /app
WORKDIR /app

# Download dependencies
COPY ui/package*.json /app/
RUN npm ci --omit=dev --ignore-scripts

# Build frontend
COPY ui /app
RUN npm run build -- --base=$UI_PUBLIC_URL

### Build API ###
FROM --platform=${BUILDPLATFORM:-linux/amd64} debian:bookworm AS api
ARG TARGETPLATFORM

COPY docker/install_build_dependencies.sh /tmp/
RUN chmod +x /tmp/install_build_dependencies.sh && /tmp/install_build_dependencies.sh

COPY docker/go_wrapper.sh /go/bin/go
RUN chmod +x /go/bin/go
ENV GOPATH="/go"
ENV PATH="${GOPATH}/bin:${PATH}"

ENV CGO_ENABLED 1

RUN go env

RUN mkdir -p /app
WORKDIR /app

# Download dependencies
COPY api/go.mod api/go.sum /app/
RUN go mod download

# Patch go-face
RUN sed -i 's/-march=native//g' ${GOPATH}/pkg/mod/github.com/!kagami/go-face*/face.go

# Build dependencies that use CGO
RUN go install \
  github.com/mattn/go-sqlite3 \
  github.com/Kagami/go-face

# Copy and build api source
COPY api /app
RUN go build -v -o photoview .

### Copy api and ui to production environment ###
FROM debian:bookworm
ARG TARGETPLATFORM
WORKDIR /app

COPY api/data /app/data

RUN apt update \
  # Required dependencies
  && apt install -y curl gpg libdlib19.1 ffmpeg exiftool libheif1

# Install Darktable if building for a supported architecture
# And create darktable directories for non-root users
RUN if [ "${TARGETPLATFORM}" = "linux/amd64" ] || [ "${TARGETPLATFORM}" = "linux/arm64" ]; then \
  apt install -y darktable && \
  mkdir -p /.cache/darktable /.config/darktable; fi

# Remove build dependencies and cleanup
RUN apt purge -y gpg \
  && apt autoremove -y \
  && apt clean \
  && rm -rf /var/lib/apt/lists/*

COPY --from=ui /app/dist /ui
COPY --from=api /app/photoview /app/photoview

ENV PHOTOVIEW_LISTEN_IP 127.0.0.1
ENV PHOTOVIEW_LISTEN_PORT 80

ENV PHOTOVIEW_SERVE_UI 1
ENV PHOTOVIEW_UI_PATH /ui

EXPOSE 80

HEALTHCHECK --interval=60s --timeout=10s CMD curl --fail 'http://localhost:80/api/graphql' -X POST -H 'Content-Type: application/json' --data-raw '{"operationName":"CheckInitialSetup","variables":{},"query":"query CheckInitialSetup { siteInfo { initialSetup }}"}'

ENTRYPOINT ["/app/photoview"]
