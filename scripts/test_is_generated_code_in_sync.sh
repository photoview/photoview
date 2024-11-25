#!/bin/sh
set -eu

cd $(dirname $0)/../api
go generate ./...
if [ "$(git status -s 2>/dev/null | head -1)" != "" ]; then
  echo '--- FAIL: The generated API code is out of sync with the recent changes. Please run `go generate ./...` under `./api` to regenerate it and commit it to this branch.'
  echo 'These are the changes:'
  git status -s
  exit 1
fi

echo '--- PASS: All generated code is in sync with the project.'
