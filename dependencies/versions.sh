#!/bin/bash

# To inject versions into current environment:
# $ export $(./versions.sh | xargs)

set -euo pipefail

: "${USER_AGENT:=}"
: "${GITHUB_TOKEN:=}"

CURL_FLAGS=(
  -fsSL
  --retry 3
  --retry-delay 5
  --retry-max-time 60
  --retry-all-errors
  --connect-timeout 10
  -H 'Accept: application/vnd.github+json'
  -H 'X-GitHub-Api-Version: 2022-11-28'
)

if [[ "${USER_AGENT}" != "" ]]; then
  CURL_FLAGS+=(-H "User-Agent: ${USER_AGENT}")
fi

if [[ "${GITHUB_TOKEN}" != "" ]]; then
  CURL_FLAGS+=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
fi

# Fetch latest version tags from GitHub releases
LIBRAW_VERSION=$(curl "${CURL_FLAGS[@]}" \
  "https://api.github.com/repos/LibRaw/LibRaw/releases/latest" | jq -r '.tag_name // ""')
IMAGEMAGICK_VERSION=$(curl "${CURL_FLAGS[@]}" \
  "https://api.github.com/repos/ImageMagick/ImageMagick/releases/latest" | jq -r '.tag_name // ""')
JELLYFIN_FFMPEG_VERSION=$(curl "${CURL_FLAGS[@]}" \
  "https://api.github.com/repos/jellyfin/jellyfin-ffmpeg/releases/latest" | jq -r '.tag_name // ""')

# Output as environment variables
echo "LIBRAW_VERSION=${LIBRAW_VERSION}"
echo "IMAGEMAGICK_VERSION=${IMAGEMAGICK_VERSION}"
echo "JELLYFIN_FFMPEG_VERSION=${JELLYFIN_FFMPEG_VERSION}"
