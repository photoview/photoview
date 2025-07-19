#!/bin/bash
set -euxo pipefail

cd "$(dirname $0)/../api"
go test ./... -v -p 1 -race \
  -cover -coverpkg=./... -coverprofile=coverage.txt -covermode=atomic \
  -database -filesystem \
  2>&1 | tee >(go-junit-report >test-api-coverage-report.xml)
