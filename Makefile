LDFLAGS += -X "main.LutraBuildTime=$(shell date -u '+%Y-%m-%d %I:%M:%S %Z')"
LDFLAGS += -X "main.LutraBuildGitHash=$(shell git rev-parse HEAD)"

OS := $(shell uname)

DATA_FILES := $(shell find conf | sed 's/ /\\ /g')

BUILD_FLAGS:=-o lutrainit -v
TAGS=
NOW=$(shell date -u '+%Y%m%d%I%M%S')
GOVET=go tool vet -composites=false -methods=false -structtags=false

export CGO_ENABLED=1
export GOOS=linux
export GOARCH=amd64

.PHONY: build clean

all: build

check: test

govet:
	$(GOVET) main.go

build:
	go build $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -tags '$(TAGS)'

build-dev: govet
	go build $(BUILD_FLAGS) -tags '$(TAGS)'

build-dev-race: govet
	go build $(BUILD_FLAGS) -race -tags '$(TAGS)'

clean:
	find . -name ".DS_Store" -delete
	go clean -i ./...

test:
	go test -cover -race ./...
