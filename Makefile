LDFLAGS += -X "main.LutraBuildTime=$(shell date -u '+%Y-%m-%d %I:%M:%S %Z')"
LDFLAGS += -X "main.LutraBuildGitHash=$(shell git rev-parse HEAD)"

OS := $(shell uname)

TAGS=
GOVET=go tool vet -composites=false -methods=false -structtags=false

.PHONY: build clean

all: build

check: test

govet:
	$(GOVET) main.go

init:
	@echo "Building init"
	cd lutrainit && GOOS=linux GOARCH=amd64 go build -o lutrainit -ldflags '$(LDFLAGS)' -tags '$(TAGS)'

ctl:
	@echo "Building client"
	cd lutractl && GOOS=linux GOARCH=amd64 go build -o lutractl -ldflags '$(LDFLAGS)' -tags '$(TAGS)'


build: init ctl

build-dev: govet
	go build -o lutrainit -v -tags '$(TAGS)'

build-dev-race: govet
	go build -o lutrainit -v -race -tags '$(TAGS)'

clean:
	find . -name ".DS_Store" -delete
	go clean -i ./...

test:
	go test -cover -race ./...
