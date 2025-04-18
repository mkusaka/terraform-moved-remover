.PHONY: test build lint clean

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOLINT=golangci-lint
BINARY_NAME=terraform-moved-remover

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd

test:
	$(GOTEST) -v ./...

lint:
	$(GOVET) ./...
	$(GOLINT) run

clean:
	$(GOCMD) clean
	rm -f $(BINARY_NAME)
