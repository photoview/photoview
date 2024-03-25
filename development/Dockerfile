# Expects /src to be mounted to the source folder (by docker-compose.yml)

FROM debian:bookworm
ARG TARGETPLATFORM
RUN mkdir -p /development
WORKDIR /development

COPY ../docker/install_build_dependencies.sh /tmp/
RUN chmod +x /tmp/install_build_dependencies.sh && /tmp/install_build_dependencies.sh

COPY ../docker/go_wrapper.sh /go/bin/go
RUN chmod +x /go/bin/go
ENV GOPATH="/go"
ENV PATH="${GOPATH}/bin:${PATH}"

ENV CGO_ENABLED 1

RUN go env

# Download dependencies
COPY ../api/go.mod ../api/go.sum /development/
RUN go mod download

# Patch go-face
RUN sed -i 's/-march=native//g' ${GOPATH}/pkg/mod/github.com/!kagami/go-face*/face.go

# Build dependencies that use CGO
RUN go install \
  github.com/mattn/go-sqlite3 \
  github.com/Kagami/go-face

# Install tools
RUN apt-get update && apt-get install -y curl procps exiftool

# Setup node
RUN curl -fsSL https://deb.nodesource.com/setup_16.x | bash -
RUN apt-get update && apt-get install -y nodejs

# Install Air for Go live reload
RUN go get -u github.com/cosmtrek/air

# Install Delve debugger
RUN go install github.com/go-delve/delve/cmd/dlv@latest

ENV PHOTOVIEW_LISTEN_IP localhost
ENV PHOTOVIEW_LISTEN_PORT 4001
ENV REACT_APP_API_ENDPOINT http://localhost:4001/
ENV PHOTOVIEW_API_ENDPOINT http://localhost:4001/
ENV PHOTOVIEW_UI_ENDPOINT http://localhost:1234/
ENV PHOTOVIEW_SERVE_UI 0
ENV PHOTOVIEW_UI_PATH /ui
ENV PHOTOVIEW_DEVELOPMENT_MODE 1
ENV COMPILER_ARCH aarch64-linux-gnu
EXPOSE 4001
EXPOSE 1234

COPY ../development/dev_startup.sh /development/
RUN chmod  +x /development/dev_startup.sh
ENTRYPOINT ["/development/dev_startup.sh"]
