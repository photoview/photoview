#!/bin/sh
set -eu

for test in "$(dirname $0)/test_*"
do
  if [ "${test}" != "${test%%/scripts/test_all.sh}" ]
  then
    continue
  fi

  echo Running ${test}...
  ${test}
done
