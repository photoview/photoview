#!/bin/bash

for db in sqlite mysql postgres; do
  echo testing against database ${db}
  ./scripts/compose.sh run -e PHOTOVIEW_DATABASE_DRIVER=${db} api go test ./... -filesystem -database -p 1 -v -covermode=atomic -coverprofile=../dev/coverage.${db}.txt
done
