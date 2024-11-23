#!/bin/sh
set -eu

go install github.com/jstemmer/go-junit-report@latest

for test in $(dirname $0)/test_*; do
  if [ "${test}" != "${test%%/scripts/test_all.sh}" ]; then
    continue
  fi

  echo Running ${test}...
  ${test}
done
