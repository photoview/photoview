#!/bin/bash

docker compose up -d --wait mysql postgres

for db in sqlite mysql postgres; do
  echo testing against database ${db}
  docker compose run -e PHOTOVIEW_DATABASE_DRIVER=${db} api go test ./... -filesystem -database -p 1 -v -covermode=atomic -coverprofile=coverage.${db}.txt
done

docker compose down
