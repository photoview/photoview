MAKEFLAGS += --always-make

-include .env
export

## Ensure compatibility with "docker-compose" (old) and "docker compose" (new).
HAS_DOCKER_COMPOSE_NO_DASH := $(shell docker compose version)
ifdef HAS_DOCKER_COMPOSE_NO_DASH
    DOCKER_COMPOSE=docker compose
else
    DOCKER_COMPOSE=docker-compose
endif
## If you want to use only the new compose command, comment previous section
## and uncomment next line:
# DOCKER_COMPOSE=docker compose

help:
	@echo 'PhotoView Docker Compose management scenarios simplification'
	@echo 'USAGE:'
	@echo 'make <target>'
	@echo ''
	@echo 'Targets:'
	@echo '   help      Prints this usage info.'
	@echo '   all       Pulls fresh Docker images from the Registry and (re)starts the service.'
	@echo '             Useful for the 1st start or update scenarios.'
	@echo '   update    The same as `all`, created for convenience.'
	@echo '   start     Creates folders for service data in the ${HOST_PHOTOVIEW_LOCATION} if not exist,'
	@echo '             and starts the service. Optionally runs a Docker system cleanup, if uncommented.'
	@echo '   stop      Just stops the service, keeping all containers and volumes in Docker.'
	@echo '   restart   Simply stops and starts the service.'
	@echo '   backup    Verifies service database and creates new service backup'
	@echo '             in the ${HOST_PHOTOVIEW_BACKUP}/<date of execution> using .tar.xz by default.'
	@echo '             If you want to use 7zz instead (which is faster), read the comment on top of the target script.'
	@echo '   pull      Pulls fresh Docker images from the Registry.'
	@echo '   terminal  Starts a Bash shell session inside the `photoview` container for troubleshooting.'
	@echo '   logs      Shows the last 100 lines (if the command not modified) from the log,'
	@echo '             stays listening for new lines, and shows them interactively. Ctrl + C to exit.'
	@echo '             It can be used as a source for a log viewer, like "make logs | lnav", so that you can'
	@echo '             interactively filter and search for needed info.'
	@echo '   down      The same as `stop`, but also removes containers and volumes from Docker. Your data is safe.'
	@echo '   remove    Removes the service from Docker, including all items.'
	@echo '   uninstall Stops and removes the service from Docker, including all items.'
	@echo ''
all: pull restart
uninstall: down remove
restart: stop start
update: pull restart
pull:
	$(DOCKER_COMPOSE) pull --ignore-pull-failures
start:
	@## The next line is for MariaDB. If you use SQLite or PostgreSQL, comment it out and uncomment the corresponding line instead
	mkdir -p ${HOST_PHOTOVIEW_LOCATION}/database/mariadb
	@## The next line is for SQLite
	@# mkdir -p ${HOST_PHOTOVIEW_LOCATION}/database
	@## The next line is for PostgreSQL
	@# mkdir -p ${HOST_PHOTOVIEW_LOCATION}/database/postgres
	mkdir -p ${HOST_PHOTOVIEW_LOCATION}/storage
	$(DOCKER_COMPOSE) up -d --remove-orphans
	@## Uncomment the next line if you want to run an automatic cleanup of Docker leftovers
	@## Make sure to read the Docker documentation to understand how it works
	@## Please note that this command is applied to the Docker host affecting all hosted services, not only the PhotoView
	@# docker system prune -f
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
## By default the `backup` uses MariaDB. If you use SQLite or PostgreSQL, replace the `backup-mariadb` with corresponding target
## keeping the order of the calls the same
## -----------------------
## To see the content of the *.tar.xz use the command `tar -tvJf archive_name.tar.xz`
## To unpack the *.tar.xz into current folder use the command `tar -xJf archive_name.tar.xz`
## -----------------------
## The backup script creates .tar.xz archives. This type of archives provides great compression rate, but utilizes a lot of
## resources and time. It was selected, because it is pre-installed on most distros.
## However, you could replace it with the 7zz, which uses much less resources with comparable compression rate.
## Make sure to install the 7zz first and then comment out the 2 lines with tar command (1st one in the corresponding DB backup
## target, 2nd one in the `backup-post` target), uncomment corresponding lines with 7zz command
backup: backup-pre backup-mariadb backup-post
backup-pre:
	mkdir -p ${HOST_PHOTOVIEW_BACKUP}
	mkdir ${HOST_PHOTOVIEW_BACKUP}/`date +%Y-%m-%d`
backup-mariadb:
	$(DOCKER_COMPOSE) exec mariadb mysqlcheck -u root --password=${MARIADB_ROOT_PASSWORD} --check --check-upgrade --flush --process-views=YES --auto-repair --all-databases
	$(DOCKER_COMPOSE) exec mariadb mysqlcheck -u root --password=${MARIADB_ROOT_PASSWORD} --optimize --flush --auto-repair --all-databases
	$(DOCKER_COMPOSE) exec mariadb mariadb-dump -u root --password=${MARIADB_ROOT_PASSWORD} -e -x --all-databases -- > ${HOST_PHOTOVIEW_BACKUP}/`date +%Y-%m-%d`/mariaDB_mysql_dump.sql
	tar -cJf ${HOST_PHOTOVIEW_BACKUP}/`date +%Y-%m-%d`/mariaDB_mysql_dump.tar.xz ${HOST_PHOTOVIEW_BACKUP}/`date +%Y-%m-%d`/mariaDB_mysql_dump.sql --remove-files
	@# 7zz a -mx=9 ${HOST_PHOTOVIEW_BACKUP}/`date +%Y-%m-%d`/mariaDB_mysql_dump.7z ${HOST_PHOTOVIEW_BACKUP}/`date +%Y-%m-%d`/mariaDB_mysql_dump.sql && rm ${HOST_PHOTOVIEW_BACKUP}/`date +%Y-%m-%d`/mariaDB_mysql_dump.sql
backup-sqlite:
	$(DOCKER_COMPOSE) exec photoview sh -c 'echo "PRAGMA integrity_check;" | sqlite3 /app/database/photoview.db'
	$(DOCKER_COMPOSE) exec photoview sh -c 'echo "VACUUM;" | sqlite3 /app/database/photoview.db'
	$(DOCKER_COMPOSE) exec photoview sh -c 'echo ".backup \"/app/database/sqlite_backup.db\"" | sqlite3 /app/database/photoview.db'
	tar -cJf ${HOST_PHOTOVIEW_BACKUP}/`date +%Y-%m-%d`/sqlite_backup.tar.xz ${HOST_PHOTOVIEW_LOCATION}/database/sqlite_backup.db --remove-files
	@# 7zz a -mx=9 ${HOST_PHOTOVIEW_BACKUP}/`date +%Y-%m-%d`/sqlite_backup.7z ${HOST_PHOTOVIEW_LOCATION}/database/sqlite_backup.db && rm ${HOST_PHOTOVIEW_LOCATION}/database/sqlite_backup.db
backup-postgres:
	$(DOCKER_COMPOSE) exec postgres PGPASSWORD=${PGSQL_PASSWORD} psql -U ${PGSQL_USER} -d ${PGSQL_DATABASE} -c "VACUUM (VERBOSE, ANALYZE)"
	$(DOCKER_COMPOSE) exec postgres PGPASSWORD=${PGSQL_PASSWORD} pg_dump -U ${PGSQL_USER} -c --if-exists -C -v ${PGSQL_DATABASE} -- > ${HOST_PHOTOVIEW_BACKUP}/`date +%Y-%m-%d`/postgres_pg_dump.sql
	tar -cJf ${HOST_PHOTOVIEW_BACKUP}/`date +%Y-%m-%d`/postgres_pg_dump.tar.xz ${HOST_PHOTOVIEW_BACKUP}/`date +%Y-%m-%d`/postgres_pg_dump.sql --remove-files
	@# 7zz a -mx=9 ${HOST_PHOTOVIEW_BACKUP}/`date +%Y-%m-%d`/postgres_pg_dump.7z ${HOST_PHOTOVIEW_BACKUP}/`date +%Y-%m-%d`/postgres_pg_dump.sql && rm ${HOST_PHOTOVIEW_BACKUP}/`date +%Y-%m-%d`/postgres_pg_dump.sql
backup-post:
	cp ${HOST_PHOTOVIEW_LOCATION}/Makefile ${HOST_PHOTOVIEW_BACKUP}/`date +%Y-%m-%d`/
	cp ${HOST_PHOTOVIEW_LOCATION}/docker-compose.yml ${HOST_PHOTOVIEW_BACKUP}/`date +%Y-%m-%d`/
	cp ${HOST_PHOTOVIEW_LOCATION}/.env ${HOST_PHOTOVIEW_BACKUP}/`date +%Y-%m-%d`/
	tar -cJf ${HOST_PHOTOVIEW_BACKUP}/`date +%Y-%m-%d`/storage.tar.xz ${HOST_PHOTOVIEW_LOCATION}/storage
	@# 7zz a -mx=9 ${HOST_PHOTOVIEW_BACKUP}/`date +%Y-%m-%d`/storage.7z ${HOST_PHOTOVIEW_LOCATION}/storage
