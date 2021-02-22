#!/bin/sh

# Script to configure environment variables for Go compiler
# to allow cross compilation

: ${TARGETPLATFORM=}
: ${TARGETOS=}
: ${TARGETARCH=}
: ${TARGETVARIANT=}
: ${CGO_ENABLED=}
: ${GOARCH=}
: ${GOOS=}
: ${GOARM=}
: ${GOBIN=}

set -eu

if [ ! -z "$TARGETPLATFORM" ]; then
  os="$(echo $TARGETPLATFORM | cut -d"/" -f1)"
  arch="$(echo $TARGETPLATFORM | cut -d"/" -f2)"
  if [ ! -z "$os" ] && [ ! -z "$arch" ]; then
    export GOOS="$os"
    export GOARCH="$arch"
    if [ "$arch" = "arm" ]; then
      case "$(echo $TARGETPLATFORM | cut -d"/" -f3)" in
      "v5")
        export GOARM="5"
        ;;
      "v6")
        export GOARM="6"
        ;;
      *)
        export GOARM="7"
        ;;
      esac
    fi
  fi
fi

if [ ! -z "$TARGETOS" ]; then
  export GOOS="$TARGETOS"
fi

if [ ! -z "$TARGETARCH" ]; then
  export GOARCH="$TARGETARCH"
fi

if [ "$TARGETARCH" = "arm" ]; then
  if [ ! -z "$TARGETVARIANT" ]; then
    case "$TARGETVARIANT" in
    "v5")
      export GOARM="5"
      ;;
    "v6")
      export GOARM="6"
      ;;
    *)
      export GOARM="7"
      ;;
    esac
  else
    export GOARM="7"
  fi
fi

if [ "$CGO_ENABLED" = "1" ]; then
  case "$GOARCH" in
  "amd64")
    export CC="x86_64-linux-gnu-gcc"
    export CXX="x86_64-linux-gnu-g++"
    ;;
  "ppc64le")
    export CC="powerpc64le-linux-gnu-gcc"
    export CXX="powerpc64le-linux-gnu-g++"
    ;;
  "s390x")
    export CC="s390x-linux-gnu-gcc"
    export CXX="s390x-linux-gnu-g++"
    ;;
  "arm64")
    export CC="aarch64-linux-gnu-gcc"
    export CXX="aarch64-linux-gnu-g++"
    ;;
  "arm")
    case "$GOARM" in
    "5")
      export CC="arm-linux-gnueabi-gcc"
      export CXX="arm-linux-gnueabi-g++"
      ;;
    *)
      export CC="arm-linux-gnueabihf-gcc"
      export CXX="arm-linux-gnueabihf-g++"
      ;;
    esac
    ;;
  esac
fi

if [ "$GOOS" = "wasi" ]; then
  export GOOS="js"
fi

if [ -z "$GOBIN" ] && [ -n "$GOPATH" ] && [ -n "$GOARCH" ] && [ -n "$GOOS" ]; then
  export PATH=${GOPATH}/bin/${GOOS}_${GOARCH}:${PATH}
fi

exec /usr/local/go/bin/go "$@"
