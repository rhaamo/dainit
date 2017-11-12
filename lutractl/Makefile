EXECUTABLE := lutractl

LDFLAGS += -X "github.com/rhaamo/lutrainit/lutractl/setting.BuildTime=$(shell date -u '+%Y-%m-%d %I:%M:%S %Z')"
LDFLAGS += -X "github.com/rhaamo/lutrainit/lutractl/setting.BuildGitHash=$(shell git rev-parse HEAD)"

GOFILES := $(shell find . -name "*.go" -type f ! -path "./vendor/*" ! -path "*/bindata.go")
PACKAGES ?= $(shell go list ./... | grep -v /vendor/)

include ../common.mk

