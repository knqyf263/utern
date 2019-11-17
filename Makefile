VERSION = $(shell godzil show-version)
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X main.commit=$(CURRENT_REVISION)"
u := $(if $(update),-u)

.PHONY: deps
deps:
	go get ${u} -d
	go mod tidy

.PHONY: devel-deps
devel-deps:
	GO111MODULE=off go get ${u} \
	  golang.org/x/lint/golint                  \
	  github.com/Songmu/godzil/cmd/godzil       \
	  github.com/Songmu/gocredits/cmd/gocredits

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint: devel-deps
	golint -set_exit_status

.PHONY: build
build:
	go build -ldflags=$(BUILD_LDFLAGS)

.PHONY: install
install:
	go install -ldflags=$(BUILD_LDFLAGS)

.PHONY: release
release: devel-deps
	godzil release

CREDITS: deps devel-deps go.sum
	gocredits -w
