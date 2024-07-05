#!/bin/bash

declare -A DBS

while (( "$#" )); do
  case "$1" in
    sqlite)
      DBS["sqlite"]="true"
      ;;

    mysql)
      DBS["mysql"]="true"
      ;;

    postgres)
      DBS["postgres"]="true"
      ;;

    all)
      DBS["sqlite"]="true"
      DBS["mysql"]="true"
      DBS["postgres"]="true"
      ;;

    --)
      shift
      break;
      ;;

    *)
      echo "$0 <all | sqlite | mysql | postgres> [-- <test arguments>]"
      exit -1
      ;;
  esac

  shift
done

if [ "${#DBS[@]}" = "0" ]; then
  DBS["sqlite"]="true"
fi

for db in ${!DBS[@]}; do
  echo testing ${db} with args: $@
  ./scripts/compose.sh run -e PHOTOVIEW_DATABASE_DRIVER=${db} api-test \
    go test ./... $@
done
