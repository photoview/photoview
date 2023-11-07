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
	@echo '   start     Creates folders for service data in the ${HOST_PHOTOVIW_LOCATION} if not exist,'
	@echo '             and starts the service. Optionally runs a Docker system cleanup, if uncommented.'
	@echo '   stop      Just stops the service, keeping all containers and volumes in Docker.'
	@echo '   restart   Simply stops and starts the service.'
	@echo '   backup    Verifies service database and creates new service backup'
	@echo '             in the ${HOST_PHOTOVIW_BACKUP}/<date of execution> using .tar.xz by default.'
	@echo '             If you want to use 7zz instead (which is faster), read the comment in the target script.'
	@echo '   pull      Pulls fresh Docker images from the Registry.'
	@echo '   terminal  Starts a Bash shell session inside the `photoview` container for troubleshooting.'
	@echo '   logs      Shows the last 100 lines (if the command not modified) from the log'
	@echo '             and stays listening for new lines and show them interactively. Ctrl + C to exit.'
	@echo '   down      The same as `stop`, but also removes containers and volumes from Docker.'
	@echo '   remove    Removes the service from Docker, including all items.'
	@echo '   uninstall Stops and removes the service from Docker, including all items.'
	@echo '   dev       Pulls, builds and (re)starts the service in development mode.'
	@echo '   dev-down  The same as `down`, but for service in development mode.'
	@echo ''
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
	## Uncomment the next line if you want to run an automatic cleanup of Docker leftovers
	## Make sure to read the Docker documentation to understand how it works
	## Please note that this command is applied to the Docker host affecting all hosted services, not only the PhotoView
	# docker system prune -f
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
	## To see the content of the *.tar.xz use the command `tar -tvJf archive_name.tar.xz`
	## To unpack the *.tar.xz into current folder use the command `tar -xJf archive_name.tar.xz`
	## -----------------------
	## The backup script creates .tar.xz archives. This type of archives provides great compression rate, but utilizes a lot of resources and time.
	## It was selected, because it is pre-installed on most distros. However, you could replace it with the 7zz, which uses much less resources with comparable compression rate.
	## Make sure to install the 7zz first and then comment out the 2 lines with tar command before this comment, uncomment the next lines
	# 7zz a -mx=9 ${HOST_PHOTOVIW_BACKUP}/`date +%Y-%m-%d`/mariaDB_mysql_dump.7z ${HOST_PHOTOVIW_BACKUP}/`date +%Y-%m-%d`/mariaDB_mysql_dump.sql
    # 7zz t ${HOST_PHOTOVIW_BACKUP}/`date +%Y-%m-%d`/mariaDB_mysql_dump.7z && rm ${HOST_PHOTOVIW_BACKUP}/`date +%Y-%m-%d`/mariaDB_mysql_dump.sql || exit 1
    # 7zz a -mx=9 ${HOST_PHOTOVIW_BACKUP}/`date +%Y-%m-%d`/storage.7z ${HOST_PHOTOVIW_LOCATION}/storage
    # 7zz t ${HOST_PHOTOVIW_BACKUP}/`date +%Y-%m-%d`/storage.7z
dev-down:
	$(DOCKER_COMPOSE) -f ../docker-compose-dev.yml down
dev: dev-down
	$(DOCKER_COMPOSE) -f ../docker-compose-dev.yml pull --ignore-pull-failures
	mkdir -p ${HOST_PHOTOVIW_LOCATION}/database
	mkdir -p ${HOST_PHOTOVIW_LOCATION}/storage
	$(DOCKER_COMPOSE) -f ../docker-compose-dev.yml build --build-arg COMMIT_SHA=$$(git rev-parse --short HEAD) --build-arg BUILD_DATE=$$(date +%Y-%m-%d) --build-arg VERSION="dev" --pull photoview-dev
	$(DOCKER_COMPOSE) -f ../docker-compose-dev.yml up -d --remove-orphans
