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

.PHONY: build
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/...

.PHONY: test
test:
	$(GOTEST) -v ./...

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

.PHONY: install
install:
	$(GOGET) -v ./...
	$(GOGET) github.com/stretchr/testify/assert


.PHONY: lint
lint:
	@hash golint > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GOGET) -u golang.org/x/lint/golint; \
	fi
	
	for PKG in $(PACKAGES); do golint -set_exit_status $$PKG || exit 1; done;

.PHONY: view-covered
view-covered:
	$(GOTEST) -coverprofile=cover.out $(TARGET)
	$(GOCMD) tool cover -html=cover.out
