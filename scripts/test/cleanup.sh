#!/bin/sh

docker compose -f "$(dirname $0)/../../dev-compose.yaml" down

