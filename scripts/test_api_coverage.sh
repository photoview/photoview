#!/bin/sh
set -eu

cd $(dirname $0)/../api
go test ./... -v -database -filesystem -p 1 -coverprofile=coverage.txt -covermode=atomic
