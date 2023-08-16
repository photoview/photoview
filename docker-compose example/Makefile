MAKEFLAGS += --always-make

-include .env
export

## Ensure compatibility with "docker-compose" (old) and "docker compose" (new).
HAS_DOCKER_COMPOSE_WITH_DASH := $(shell which docker-compose)
ifdef HAS_DOCKER_COMPOSE_WITH_DASH
    DOCKER_COMPOSE=docker-compose
else
    DOCKER_COMPOSE=docker compose
endif
## If you want to use only the new compose command, comment previous section
## and uncomment next line:
# DOCKER_COMPOSE=docker compose

all: pull restart
uninstall: down remove
restart: stop start
update: pull restart
pull:
	$(DOCKER_COMPOSE) pull --ignore-pull-failures
start:
	mkdir -p ${HOST_PHOTOVIW_LOCATION}/database
	mkdir -p ${HOST_PHOTOVIW_LOCATION}/storage
	$(DOCKER_COMPOSE) up -d --remove-orphans
	docker system prune -f
stop:
	$(DOCKER_COMPOSE) stop
down:
	$(DOCKER_COMPOSE) down -v
remove:
	$(DOCKER_COMPOSE) rm -s -v
terminal:
	$(DOCKER_COMPOSE) exec photoview bash
logs:
	$(DOCKER_COMPOSE) logs --tail=100 -f
backup:
	$(DOCKER_COMPOSE) exec db mysqlcheck -u root --password=${MARIADB_ROOT_PASSWORD} --check --check-upgrade --flush --process-views=YES --auto-repair --all-databases
	$(DOCKER_COMPOSE) exec db mysqlcheck -u root --password=${MARIADB_ROOT_PASSWORD} --optimize --flush --auto-repair --all-databases
	mkdir -p ${HOST_PHOTOVIW_BACKUP}
	mkdir ${HOST_PHOTOVIW_BACKUP}/`date +%Y-%m-%d`
	cp ${HOST_PHOTOVIW_LOCATION}/Makefile ${HOST_PHOTOVIW_BACKUP}/`date +%Y-%m-%d`/
	cp ${HOST_PHOTOVIW_LOCATION}/docker-compose.yml ${HOST_PHOTOVIW_BACKUP}/`date +%Y-%m-%d`/
	cp ${HOST_PHOTOVIW_LOCATION}/.env ${HOST_PHOTOVIW_BACKUP}/`date +%Y-%m-%d`/
	$(DOCKER_COMPOSE) exec db mariadb-dump -u root --password=${MARIADB_ROOT_PASSWORD} -e -x --all-databases -- > ${HOST_PHOTOVIW_BACKUP}/`date +%Y-%m-%d`/mariaDB_mysql_dump.sql
	tar -cJf ${HOST_PHOTOVIW_BACKUP}/`date +%Y-%m-%d`/mariaDB_mysql_dump.tar.xz ${HOST_PHOTOVIW_BACKUP}/`date +%Y-%m-%d`/mariaDB_mysql_dump.sql --remove-files
	tar -cJf ${HOST_PHOTOVIW_BACKUP}/`date +%Y-%m-%d`/storage.tar.xz ${HOST_PHOTOVIW_LOCATION}/storage
	# To see the content of the *.tar.xz use the command `tar -tvJf archive_name.tar.xz`
	# To unpack the *.tar.xz into current folder use the command `tar -xJf archive_name.tar.xz`
dev-down:
	$(DOCKER_COMPOSE) -f ../docker-compose-dev.yml down
dev: dev-down
	$(DOCKER_COMPOSE) -f ../docker-compose-dev.yml pull --ignore-pull-failures
	mkdir -p ${HOST_PHOTOVIW_LOCATION}/database
	mkdir -p ${HOST_PHOTOVIW_LOCATION}/storage
	$(DOCKER_COMPOSE) -f ../docker-compose-dev.yml build --build-arg COMMIT_SHA=$$(git rev-parse --short HEAD) --build-arg BUILD_DATE=$$(date +%Y-%m-%d) --build-arg VERSION="dev" --pull photoview-dev
	$(DOCKER_COMPOSE) -f ../docker-compose-dev.yml up -d --remove-orphans
