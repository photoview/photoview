#!/bin/bash

declare -A DBS

for var in "$@"; do
  case "$var" in
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

    *)
      echo "$0 <all | sqlite | mysql | postgres>"
      exit -1
      ;;
  esac
done

if [ "$#" = "0" ]; then
  DBS["sqlite"]="true"
fi

for db in ${!DBS[@]}; do
  ./scripts/compose.sh run -e PHOTOVIEW_DATABASE_DRIVER=${db} api go test ./... -filesystem -database -p 1 -v
done
