#! /usr/bin/make
GOCMD=$(shell which go)
GOLINT=$(shell which golint)
GOIMPORT=$(shell which goimports)
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOLIST=$(GOCMD) list
BINARY_NAME=swag

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/...

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

DIRS=$(shell $(GOLIST) -f {{.Dir}} ./...)
lint:
	@for d in $(DIRS) ; do \
		if [ "`$(GOIMPORT) -l $$d/*.go | tee /dev/stderr`" ]; then \
			echo "^ - Repo contains improperly formatted go files" && echo && exit 1; \
		fi \
	done
	@if [ "`$(GOLINT) ./... | grep -vf .golint_exclude | tee /dev/stderr`" ]; then \
		echo "^ - Lint errors!" && echo && exit 1; \
	fi

deps:
	$(GOGET) -v ./...
	$(GOGET) github.com/stretchr/testify/assert
	$(GOGET) github.com/golang/lint/golint
	$(GOGET) golang.org/x/tools/cmd/goimports
