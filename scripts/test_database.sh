#!/bin/bash

declare -A DBS
declare -A DEPS

for var in "$@"; do
  case "$var" in
    sqlite)
      DBS["sqlite"]="true"
      ;;

    mysql)
      DEPS["mysql"]="true"
      DBS["mysql"]="true"
      ;;

    postgres)
      DEPS["postgres"]="true"
      DBS["postgres"]="true"
      ;;

    all)
      DBS["sqlite"]="true"
      DEPS["mysql"]="true"
      DBS["mysql"]="true"
      DEPS["postgres"]="true"
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

if [ "${#DEPS[@]}" -ne "0" ]; then
  docker compose up -d --wait ${!DEPS[@]}
fi

for db in ${!DBS[@]}; do
  docker compose run -e PHOTOVIEW_DATABASE_DRIVER=${db} api go test ./... -filesystem -database -p 1 -v
done

if [ "${#DEPS[@]}" -ne "0" ]; then
  docker compose down ${!DEPS[@]}
fi
