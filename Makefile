SHELL := /bin/bash

define load_env
	@set -a && . $(1) && set +a && \
	$(MAKE) --no-print-directory $(2)
endef

# application binary type for docker image
export DOCKER_GOOS=linux
export DOCKER_GOARCH=amd64
# Go version to use while building binaries for docker image
export GOLANG_VERSION=1.24
# golang OS tag for building binaries for docker image
export GOLANG_IMAGE=alpine 
# target OS: the image type to run in production. Usually alpine fits OK.
export TARGET_DISTR_TYPE=alpine
# target OS version (codename)
export TARGET_DISTR_VERSION=latest
# a user created inside the fcache container
# files created by those services on mounted volumes will be owned by this user
export DOCKER_USER=$(USER)

LDFLAGS = -s -w -X main.appVersion=dev-$(shell git rev-parse --short HEAD)-$(shell date +%y-%m-%d-%H%M%S)
PROJECT = $(shell basename $(PWD))
BIN = ./bin
SRC = ./cmd
BINARY = $(BIN)/$(PROJECT)

define USAGE

Usage: make <target>

some of the <targets> are:
  setup                - Set up the project
  update-deps          - update Go dependencies
  all                  - build + lint + gosec + test
  build                - build binaries into $(BIN)/
  lint                 - run linters
  gosec                - Go security checker
  test                 - run tests
  docker-build         - build docker images
  db-up                - run the test DB for local development
  db-down              - stop the test DB
  reset-db             - Reset the test DB
  {dev|stage}-up       - run the app in developer | staging mode
  down                 - stop the app

endef
export USAGE

define SETUP_HELP

This command will set up the project on the local machine.

What it will do:
    - install dependencies and tools (linter)
    - create local directories for temporary and cache files

Press Enter to continue, Ctrl+C to quit
endef
export SETUP_HELP

define DOCKER_PARAMS
--build-arg USER=$(DOCKER_USER) \
--build-arg GOOS=$(DOCKER_GOOS) \
--build-arg GOARCH=$(DOCKER_GOARCH) \
--build-arg GOLANG_VERSION=$(GOLANG_VERSION) \
--build-arg GOLANG_IMAGE=$(GOLANG_IMAGE) \
--build-arg TARGET_DISTR_TYPE=$(TARGET_DISTR_TYPE) \
--build-arg TARGET_DISTR_VERSION=$(TARGET_DISTR_VERSION) \
--build-arg LDFLAGS="$(LDFLAGS)" \
--file Dockerfile
endef
export DOCKER_PARAMS

define CAKE
   \033[1;31m. . .\033[0m
   i i i
  %~%~%~%
  |||||||
-=========-
endef
export CAKE

help:
	@echo "$$USAGE"

setup:
	@echo "$$SETUP_HELP"
	read
	go install github.com/golangci/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest

update-deps:
	go get -u ./... && go mod tidy && go mod vendor

all: build lint gosec test

tidy:
	go mod tidy
	go mod vendor

build:
	mkdir -p $(BIN)
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -trimpath -o $(BIN)/fcache $(SRC)

lint:
	golangci-lint run

test:
	go test ./...	

run:
	CGO_ENABLED=0 go run -ldflags "$(LDFLAGS)" -trimpath $(SRC) --config-file=config/config.yaml

docker-image:
	docker build --tag fcache --target fcache $(DOCKER_PARAMS) .

db-up:
	$(call load_env,.env_dev,db-up-run)

db-up-run:
	docker compose -f docker-compose.dev.db.yaml up -d
	echo "DB available at ${CACHE_DB_PORT}"

db-down:
	docker compose -f docker-compose.dev.db.yaml down

reset-db:
	docker volume rm fcache_db_data

dev-up: tidy
	echo "Running fcache locally"
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -trimpath -o $(BIN)/fcache $(SRC)
	./$(BIN)/fcache --config-file=config/config.yaml --print-config=true > $(BIN)/fcache.log 2>&1 & echo $$! > $(BIN)/fcache.pid
	echo Fcache PID: `cat $(BIN)/fcache.pid`

stage-up: export BUILD_MODE=stage
stage-up: tidy
	$(call load_env,.env_stage,run-up)

run-up: export CACHE_UPSTREAM=fcache
run-up: docker-build
	mkdir -p /var/nginx-proxy/domains/$(CACHE_DOMAIN)/cache
	docker compose -f docker-compose.$(BUILD_MODE).yaml up -d
	sleep 0.5
	docker exec nginx-proxy cat /app/scripts/install-nginx-config.sh | bash -s -- fcache "docker-assets/nginx/*.*"
	[[ "$(BUILD_MODE)" != "prod" ]] && docker network connect --alias combobox.com combobox_net fcache-fcache-1

down:
	-kill $$(cat $(BIN)/fcache.pid) && rm $(BIN)/fcache.pid
	docker compose -f docker-compose.stage.yaml down
	docker exec nginx-proxy cat /app/scripts/remove-nginx-config.sh | bash -s -- fcache fcache.conf

cake:
	printf "%b\n" "$$CAKE"

.PHONY: all build run lint test gosec docker-image update-deps db-up db-down dev-up stage-up prod-up down cake

$(V).SILENT:
