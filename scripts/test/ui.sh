#!/bin/sh

docker compose -f "$(dirname $0)/../../test-compose.yaml" run ui
