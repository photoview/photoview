### Build UI ###
FROM --platform=${BUILDPLATFORM:-linux/amd64} node:18 AS ui
# See for details: https://github.com/hadolint/hadolint/wiki/DL4006
SHELL ["/bin/bash", "-euo", "pipefail", "-c"]

ARG NODE_ENV=production
ENV NODE_ENV=${NODE_ENV}

WORKDIR /app/ui

COPY ui/package.json ui/package-lock.json /app/ui/
RUN if [ "$NODE_ENV" = "production" ]; then \
    echo "Installing production dependencies only..."; \
    npm ci --omit=dev; \
    else \
    echo "Installing all dependencies..."; \
    npm ci; \
    fi

COPY ui/ /app/ui

# Set environment variable REACT_APP_API_ENDPOINT from build args, uses "<web server>/api" as default
ARG REACT_APP_API_ENDPOINT
ENV REACT_APP_API_ENDPOINT=${REACT_APP_API_ENDPOINT}

# Set environment variable UI_PUBLIC_URL from build args, uses "<web server>/" as default
ARG UI_PUBLIC_URL=/
ENV UI_PUBLIC_URL=${UI_PUBLIC_URL}

ARG VERSION=unknown-branch
ARG TARGETARCH
# hadolint ignore=SC2155
RUN export BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%S+00:00(UTC)'); \
    export REACT_APP_BUILD_DATE=${BUILD_DATE}; \
    export COMMIT_SHA="-=<GitHub-CI-commit-sha-placeholder>=-"; \
    export REACT_APP_BUILD_COMMIT_SHA=${COMMIT_SHA}; \
    export VERSION="${VERSION}-${TARGETARCH}"; \
    export REACT_APP_BUILD_VERSION=${VERSION}; \
    npm run build -- --base="${UI_PUBLIC_URL}"

### Build API ###
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.24-bookworm AS api
ARG TARGETPLATFORM

# See for details: https://github.com/hadolint/hadolint/wiki/DL4006
SHELL ["/bin/bash", "-euo", "pipefail", "-c"]

WORKDIR /app/api

ENV GOPATH="/go"
ENV PATH="${GOPATH}/bin:${PATH}"
ENV CGO_ENABLED=1

# Download dependencies
COPY scripts/set_compiler_env.sh /app/scripts/
RUN chmod +x /app/scripts/*.sh \
    && source /app/scripts/set_compiler_env.sh

COPY scripts/install_*.sh /app/scripts/
# Split values in `/env`
# hadolint ignore=SC2046
RUN chmod +x /app/scripts/*.sh \
    && export $(cat /env) \
    && /app/scripts/install_build_dependencies.sh \
    && /app/scripts/install_runtime_dependencies.sh

COPY --from=photoview/dependencies:latest /artifacts.tar.gz /dependencies/
# Split values in `/env`
# hadolint ignore=SC2046
RUN export $(cat /env) \
    && git config --global --add safe.directory /app \
    && cd /dependencies/ \
    && tar xfv artifacts.tar.gz \
    && cp -a include/* /usr/local/include/ \
    && cp -a pkgconfig/* ${PKG_CONFIG_PATH} \
    && cp -a lib/* /usr/local/lib/ \
    && ldconfig \
    && apt-get install -y ./deb/jellyfin-ffmpeg.deb

COPY api/go.mod api/go.sum /app/api/
# Split values in `/env`
# hadolint ignore=SC2046
RUN export $(cat /env) \
    && go env \
    && go mod download \
    # Patch go-face
    && sed -i 's/-march=native//g' ${GOPATH}/pkg/mod/github.com/!kagami/go-face*/face.go \
    # Build dependencies that use CGO
    && go install \
    github.com/mattn/go-sqlite3 \
    github.com/Kagami/go-face

COPY api /app/api
# Split values in `/env`
# hadolint ignore=SC2046
RUN export $(cat /env) \
    && go env \
    && go build -v -o photoview .

### Build release image ###
FROM debian:bookworm-slim AS release
ARG TARGETPLATFORM

# See for details: https://github.com/hadolint/hadolint/wiki/DL4006
SHELL ["/bin/bash", "-euo", "pipefail", "-c"]

COPY scripts/install_runtime_dependencies.sh /app/scripts/
RUN --mount=type=bind,from=api,source=/dependencies/,target=/dependencies/ \
    chmod +x /app/scripts/install_runtime_dependencies.sh \
    # Create a user to run Photoview server
    && groupadd -g 999 photoview \
    && useradd -r -u 999 -g photoview -m photoview \
    # Install required dependencies
    && /app/scripts/install_runtime_dependencies.sh \
    # Install self-building libs
    && cd /dependencies \
    && cp -a lib/*.so* /usr/local/lib/ \
    && ldconfig \
    && apt-get install -y ./deb/jellyfin-ffmpeg.deb \
    && ln -s /usr/lib/jellyfin-ffmpeg/ffmpeg /usr/local/bin/ \
    && ln -s /usr/lib/jellyfin-ffmpeg/ffprobe /usr/local/bin/ \
    # Cleanup
    && apt-get autoremove -y \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY api/data /app/data
COPY --from=ui /app/ui/dist /app/ui
COPY --from=api /app/api/photoview /app/photoview
# This is a w/a for letting the UI build stage to be cached
# and not rebuilt every new commit because of the build_arg value change.
ARG COMMIT_SHA=?commit?
RUN find /app/ui/assets -type f -name "SettingsPage.*.js" \
    -exec sed -i "s/=\"-=<GitHub-CI-commit-sha-placeholder>=-\";/=\"${COMMIT_SHA}\";/g" {} \;

WORKDIR /home/photoview

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
