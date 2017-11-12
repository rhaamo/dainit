OS := $(shell uname)

DIST := dist

BUILD_FLAGS := -o $(EXECUTABLE) -v
NOW=$(shell date -u '+%Y%m%d%I%M%S')

TAGS = 

GOVET=go vet
GOLINT=golint -set_exit_status
GO ?= go
GOFMT ?= gofmt -s


ifneq ($(DRONE_TAG),)
	VERSION ?= $(subst v,,$(DRONE_TAG))
else
	ifneq ($(DRONE_BRANCH),)
		VERSION ?= $(subst release/v,,$(DRONE_BRANCH))
	else
		VERSION ?= master
	endif
endif

### Targets

.PHONY: build clean

all: build

check: test

vet:
	$(GOVET) $(PACKAGES)

lint:
	@hash golint > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/golang/lint/golint; \
	fi
	for PKG in $(PACKAGES); do golint -set_exit_status $$PKG || exit 1; done;

build:
	$(GO) build $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -tags '$(TAGS)'

build-race:
	$(GO) build $(BUILD_FLAGS) -race -ldflags '$(LDFLAGS)' -tags '$(TAGS)'

clean:
	$(GO) clean -i ./...
	find . -name ".DS_Store" -delete

test: fmt-check
	$(GO) test -cover -v$(PACKAGES)

.PHONY: misspell-check
misspell-check:
	@hash misspell > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/client9/misspell/cmd/misspell; \
	fi
	misspell -error -i unknwon $(GOFILES)

.PHONY: misspell
misspell:
	@hash misspell > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) get -u github.com/client9/misspell/cmd/misspell; \
	fi
	misspell -w -i unknwon $(GOFILES)

required-gofmt-version:
	@go version  | grep -q '\(1.7\|1.8\|1.9\)' || { echo "We require go version 1.7, 1.8 or 1.9 to format code" >&2 && exit 1; }

.PHONY: fmt
fmt: required-gofmt-version
	$(GOFMT) -w $(GOFILES)

.PHONY: fmt-check
fmt-check: required-gofmt-version
	# get all go files and run go fmt on them
	@diff=$$($(GOFMT) -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

