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
PACKAGES=$(shell $(GOLIST) -f {{.Dir}} ./... | grep -v /example)

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/...

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

lint:
	@hash golint > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GOTEST) -u golang.org/x/lint/golint; \
	fi
	
	for PKG in $(PACKAGES); do golint -set_exit_status $$PKG || exit 1; done;

view-covered:
	$(GOTEST) -coverprofile=cover.out $(TARGET)
	$(GOCMD) tool cover -html=cover.out
