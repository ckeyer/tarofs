PWD := $(shell pwd)
APP := tarofs
PKG := github.com/ckeyer/$(APP)

GO := go
HASH := $(shell which shasum || which sha1sum)

VERSION := $(shell cat VERSION.txt)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_AT := $(shell date "+%Y-%m-%dT%H:%M:%SZ%z")

LD_FLAGS := -X ${PKG}/version.version=$(VERSION) \
 -X ${PKG}/version.gitCommit=$(GIT_COMMIT) \
 -X ${PKG}/version.buildAt=$(BUILD_AT)

IMAGE := ckeyer/${APP}
GO_IMAGE := ckeyer/dev:fuse

local:
	$(GO) build -v -ldflags="$(LD_FLAGS)" -o bundles/$(APP) main.go
	$(HASH) bundles/$(APP)

build: local

run: build
	# -fusermount -u /tmp/tarofs
	# rm -rf /data/tarofs_data
	bundles/$(APP) -D m

test:
	$(GO) test -v -cover -covermode=count $$(go list ./... |grep -v "vendor")

integration-test:
	$(GO) test -v -cover -covermode=count ./tests/

test-in-docker:
	docker run --rm -it \
	 --name ${APP}-dev \
	 --device /dev/fuse \
	 --cap-add SYS_ADMIN \
	 -v ${PWD}:/go/src/${PKG} \
	 -w /go/src/${PKG} \
	 ${GO_IMAGE} make test

image:
	docker build -t ${IMAGE}:${VERSION} .

clean:
	rm -rf bundles/*

dev:
	docker run --rm -it \
	 --name ${APP}-dev \
	 --device /dev/fuse \
	 --cap-add SYS_ADMIN \
	 -v ${PWD}:/go/src/${PKG} \
	 -w /go/src/${PKG} \
	 ${GO_IMAGE} bash
