#!/bin/bash
set -euo pipefail

go install github.com/jstemmer/go-junit-report@latest

for test in "$(dirname "$0")"/test_*; do
  if [[ "${test}" != "${test%%/test_all.sh}" ]]; then
    continue
  fi

  echo "Running ${test}..."
  "${test}"
done
