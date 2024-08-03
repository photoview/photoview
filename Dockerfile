### Build UI ###
FROM --platform=${BUILDPLATFORM:-linux/amd64} node:18 AS ui

# See for details: https://github.com/hadolint/hadolint/wiki/DL4006
SHELL ["/bin/bash", "-euo", "pipefail", "-c"]

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

WORKDIR /app/ui

COPY ui/package.json ui/package-lock.json /app/ui
RUN npm ci

COPY ui/ /app/ui
RUN npm run build -- --base=$UI_PUBLIC_URL

### Build API ###
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.22-bookworm AS api
ARG TARGETPLATFORM

# See for details: https://github.com/hadolint/hadolint/wiki/DL4006
SHELL ["/bin/bash", "-euo", "pipefail", "-c"]

WORKDIR /app/api

ENV GOPATH="/go"
ENV PATH="${GOPATH}/bin:${PATH}"
ENV CGO_ENABLED=1

# Download dependencies
COPY scripts/*.sh /app/scripts/
RUN chmod +x /app/scripts/*.sh \
  && /app/scripts/install_build_dependencies.sh \
  && /app/scripts/install_runtime_dependencies.sh

COPY api/go.mod api/go.sum /app/api/
RUN source /app/scripts/set_compiler_env.sh \
  && go env \
  && go mod download \
  # Patch go-face
  && sed -i 's/-march=native//g' ${GOPATH}/pkg/mod/github.com/!kagami/go-face*/face.go \
  # Build dependencies that use CGO
  && go install \
    github.com/mattn/go-sqlite3 \
    github.com/Kagami/go-face

COPY api /app/api
RUN source /app/scripts/set_compiler_env.sh \
  && go build -v -o photoview .

### Build release image ###
FROM --platform=${BUILDPLATFORM:-linux/amd64} debian:bookworm-slim AS release
ARG TARGETPLATFORM

# See for details: https://github.com/hadolint/hadolint/wiki/DL4006
SHELL ["/bin/bash", "-euo", "pipefail", "-c"]

COPY scripts/install_runtime_dependencies.sh /app/scripts/
RUN chmod +x /app/scripts/install_runtime_dependencies.sh \
  # Create a user to run Photoview server
  && groupadd -g 999 photoview \
  && useradd -r -u 999 -g photoview -m photoview \
  # Required dependencies
  && /app/scripts/install_runtime_dependencies.sh

WORKDIR /home/photoview

COPY api/data /app/data
COPY --from=ui /app/ui/dist /app/ui
COPY --from=api /app/api/photoview /app/photoview

ENV PHOTOVIEW_LISTEN_IP=127.0.0.1
ENV PHOTOVIEW_LISTEN_PORT=80

ENV PHOTOVIEW_SERVE_UI=1
ENV PHOTOVIEW_UI_PATH=/app/ui
ENV PHOTOVIEW_FACE_RECOGNITION_MODELS_PATH=/app/data/models
ENV PHOTOVIEW_MEDIA_CACHE=/home/photoview/media-cache

EXPOSE ${PHOTOVIEW_LISTEN_PORT}

HEALTHCHECK --interval=60s --timeout=10s \
  CMD curl --fail http://localhost:${PHOTOVIEW_LISTEN_PORT}/api/graphql \
    -X POST -H 'Content-Type: application/json' \
    --data-raw '{"operationName":"CheckInitialSetup","variables":{},"query":"query CheckInitialSetup { siteInfo { initialSetup }}"}' \
    || exit 1

USER photoview
ENTRYPOINT ["/app/photoview"]
