#!/bin/sh

DB=$1
if [ "$DB" = "" ]; then
  DB=sqlite
fi

docker compose -f "$(dirname $0)/../../test-compose.yaml" run api-$DB
