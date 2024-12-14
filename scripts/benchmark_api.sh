#!/bin/bash
set -euxo pipefail

cd "$(dirname $0)/../api"
go test ./... -bench . -benchmem
