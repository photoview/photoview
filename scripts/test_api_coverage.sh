#!/bin/bash
set -euxo pipeline

cd $(dirname $0)/../api
go test ./... -v -database -filesystem -p 1 -coverprofile=coverage.txt -covermode=atomic 2>&1 | tee >(go-junit-report >test-api-coverage-report.xml)
