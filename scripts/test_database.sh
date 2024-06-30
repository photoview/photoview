#!/bin/sh

case "$1" in
  "" | sqlite)
    docker compose run -e PHOTOVIEW_DATABASE_DRIVER=sqlite   api go test ./... --database -v
    ;;

  mysql)
    docker compose up -d --wait mariadb
    docker compose run -e PHOTOVIEW_DATABASE_DRIVER=mysql    api go test ./... --database -v
    ;;

  postgres)
    docker compose up -d --wait postgres
    docker compose run -e PHOTOVIEW_DATABASE_DRIVER=postgres api go test ./... --database -v
    ;;

  all)
    docker compose up -d --wait mariadb postgres
    docker compose run -e PHOTOVIEW_DATABASE_DRIVER=sqlite   api go test ./... --database -v
    docker compose run -e PHOTOVIEW_DATABASE_DRIVER=mysql    api go test ./... --database -v
    docker compose run -e PHOTOVIEW_DATABASE_DRIVER=postgres api go test ./... --database -v
    ;;

  *)
    echo "$0 <all | sqlite | mysql | postgres>"
    ;;
esac

docker compose down mariadb postgres
