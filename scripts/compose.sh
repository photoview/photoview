#!/bin/bash

docker compose -f "$(dirname $(readlink -f "$0"))/dev-compose.yaml" $@
