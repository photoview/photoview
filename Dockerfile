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
WORKDIR /app

# Create a user to run Photoview server
RUN groupadd -r photoview \
  && useradd -r -g photoview photoview \
  && mkdir -p /home/photoview \
  && chown -R photoview:photoview /app /home/photoview \
  # Required dependencies
  && apt update \
  && apt install -y curl gpg libdlib19.1 ffmpeg exiftool libheif1 \
  # Install Darktable if building for a supported architecture
  && if [ "${TARGETPLATFORM}" = "linux/amd64" ] || [ "${TARGETPLATFORM}" = "linux/arm64" ]; then \
    apt install -y darktable; \
  fi \
  # Remove build dependencies and cleanup
  && apt purge -y gpg \
  && apt autoremove -y \
  && apt clean \
  && rm -rf /var/lib/apt/lists/*

COPY --chown=photoveiw:photoveiw api/data /app/data
COPY --chown=photoveiw:photoveiw --from=ui /app/dist /ui
COPY --chown=photoveiw:photoveiw --from=api /app/photoview /app/photoview

ENV PHOTOVIEW_LISTEN_IP 127.0.0.1
ENV PHOTOVIEW_LISTEN_PORT 80

ENV PHOTOVIEW_SERVE_UI 1
ENV PHOTOVIEW_UI_PATH /ui

EXPOSE 80

HEALTHCHECK --interval=60s --timeout=10s CMD curl --fail 'http://localhost:80/api/graphql' -X POST -H 'Content-Type: application/json' --data-raw '{"operationName":"CheckInitialSetup","variables":{},"query":"query CheckInitialSetup { siteInfo { initialSetup }}"}'

USER photoview
ENTRYPOINT ["/app/photoview"]
