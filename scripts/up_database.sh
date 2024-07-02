#!/bin/bash

mkdir -p dev/{mysql,postgres}
chmod -R 777 dev/*

./scripts/compose.sh up --wait -d mysql postgres
