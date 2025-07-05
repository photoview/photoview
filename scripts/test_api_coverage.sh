#!/bin/bash
set -euxo pipefail

cd "$(dirname $0)/../api"
go test ./... -v -database -filesystem -p 1 -race -coverpkg=./... -coverprofile=test-api-coverage-report.txt -covermode=atomic
