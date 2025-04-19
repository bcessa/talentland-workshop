.PHONY: *
.DEFAULT_GOAL:=help

# Project setup
BINARY_NAME=echoctl
OWNER=bcessa
REPO=echo-service
PROJECT_REPO=github.com/$(OWNER)/$(REPO)
DOCKER_IMAGE=ghcr.io/$(OWNER)/$(REPO)
MAINTAINERS='Ben Cessa <ben@pixative.com>'

# State values
GIT_COMMIT_DATE=$(shell TZ=UTC git log -n1 --pretty=format:'%cd' --date='format-local:%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT_HASH=$(shell git log -n1 --pretty=format:'%H')
GIT_TAG=$(shell git describe --tags --always --abbrev=0 | cut -c 1-7)

# Linker tags
# https://golang.org/cmd/link/
LD_FLAGS += -s -w

# For commands that require a specific package path, default to all local
# subdirectories if no value is provided.
pkg?="..."

help:
	@echo "Commands available"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /' | sort

## build: Build for the current architecture in use, intended for development
build:
	# Build CLI application
	go build -v -ldflags '$(LD_FLAGS)' -o $(BINARY_NAME) ./cli

## build-for: Build the available binaries for the specified 'os' and 'arch'
# make build-for os=linux arch=amd64
build-for:
	CGO_ENABLED=0 GOOS=$(os) GOARCH=$(arch) \
	go build -v -ldflags '$(LD_FLAGS)' \
	-o $(BINARY_NAME)_$(os)_$(arch)$(suffix) \
	./cli

## deps: Download and compile all dependencies and intermediary products
deps:
	go mod tidy
	go clean

## docker-build: Build docker image
# https://github.com/opencontainers/image-spec/blob/master/annotations.md
docker-build:
	make build-for os=linux arch=arm64
	mv $(BINARY_NAME)_linux_arm64 $(BINARY_NAME)
	@-docker rmi $(DOCKER_IMAGE):$(GIT_TAG:v%=%)
	@docker build \
	"--label=org.opencontainers.image.title=$(BINARY_NAME)" \
	"--label=org.opencontainers.image.authors=$(MAINTAINERS)" \
	"--label=org.opencontainers.image.created=$(GIT_COMMIT_DATE)" \
	"--label=org.opencontainers.image.revision=$(GIT_COMMIT_HASH)" \
	"--label=org.opencontainers.image.version=$(GIT_TAG:v%=%)" \
	"--build-arg=GOMOD=$(shell dirname `go env GOMOD`)" \
	--rm -t $(DOCKER_IMAGE):$(GIT_TAG:v%=%) .
	@docker tag $(DOCKER_IMAGE):$(GIT_TAG:v%=%) $(DOCKER_IMAGE):latest
	@rm $(BINARY_NAME)

## docker-run: Run server with docker
docker-run:
	@docker run -it --rm -p 9090:9090 ghcr.io/bcessa/echo-service server --config=/root/config.yaml

## debugger-build: Build docker image with debugger
debugger-build:
	@docker build -f Dockerfile.delve \
	--rm -t $(DOCKER_IMAGE)-debug:$(GIT_TAG:v%=%) .
	@docker tag $(DOCKER_IMAGE)-debug:$(GIT_TAG:v%=%) $(DOCKER_IMAGE)-debug:latest

## debugger-run: Run debuggable server
debugger-run:
	@docker run -it --rm -p 9090:9090 -p 2345:2345 --name sample-server ghcr.io/bcessa/echo-service-debug:latest

## install: Install the binary to GOPATH and keep cached all compiled artifacts
install:
	@go build -v -ldflags '$(LD_FLAGS)' -o ${GOPATH}/bin/$(BINARY_NAME) ./cli
