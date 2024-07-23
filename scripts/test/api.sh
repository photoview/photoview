#!/bin/sh

DB=$1
if [ "$DB" = "" ]; then
  DB=sqlite
fi

docker compose -f "$(dirname $0)/../../dev-compose.yaml" run test-api-$DB
