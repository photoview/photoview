#!/bin/sh
set -eu

cd api
go generate ./...
if [ "$(git status -s 2>/dev/null | head -1)" != "" ]; then
  echo "Found old file(s), please run 'go generate ./...' to update them."
  exit 1
fi
