EXECUTABLE := lutrainit

LDFLAGS += -X "dev.sigpipe.me/dashie/lutrainit/lutrainit/setting.BuildTime=$(shell date -u '+%Y-%m-%d %I:%M:%S %Z')"
LDFLAGS += -X "dev.sigpipe.me/dashie/lutrainit/lutrainit/setting.BuildGitHash=$(shell git rev-parse HEAD)"

GOFILES := $(shell find . -name "*.go" -type f ! -path "./vendor/*" ! -path "*/bindata.go")
PACKAGES ?= $(shell go list ./... | grep -v /vendor/)

include ../common.mk

