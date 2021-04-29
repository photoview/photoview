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
    export COMPILER_ARCH="x86_64-linux-musl"
    ;;
  "arm64")
    export COMPILER_ARCH="aarch64-linux-musl"
    ;;
  "arm")
    export COMPILER_ARCH="arm-linux-musleabihf"
    ;;
  esac
fi

export CC="/${COMPILER_ARCH}-cross/bin/${COMPILER_ARCH}-gcc"
export CXX="/${COMPILER_ARCH}-cross/bin/${COMPILER_ARCH}-g++"

export CGO_CPPFLAGS="-I/${COMPILER_ARCH}-cross/include -Wno-error=parentheses"
export LD_LIBRARY_PATH="/${COMPILER_ARCH}-cross/lib"
export LIBRARY_PATH="/${COMPILER_ARCH}-cross/lib"
export CGO_LDFLAGS="-g -O2 -L/${COMPILER_ARCH}-cross/lib -Wl,-rpath-link=/${COMPILER_ARCH}-cross/lib"
export LDFLAGS="-g -O2 -L/${COMPILER_ARCH}-cross/lib -Wl,-rpath-link=/${COMPILER_ARCH}-cross/lib"

export PATH="/${COMPILER_ARCH}-cross/bin:${PATH}"
export PKG_CONFIG_PATH="/${COMPILER_ARCH}-cross/usr/lib/pkgconfig/"

# /usr/bin/go env; exit 1

if [ -z "$GOBIN" ] && [ -n "$GOPATH" ] && [ -n "$GOARCH" ] && [ -n "$GOOS" ]; then
  export PATH=${GOPATH}/bin/${GOOS}_${GOARCH}:${PATH}
fi

exec /usr/bin/go "$@"
