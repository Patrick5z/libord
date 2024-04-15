SHELL:=/bin/sh
.PHONY: ord-indexer
.PHONY: ord-validator

GOCMD=go
GOBUILD=$(GOCMD) build -trimpath
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOBIN=./build

clean:
	$(GOCLEAN)
	rm -fr $(GOBIN)/*

ord-indexer:
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(GOBIN)/ord-indexer cmd/indexer/main.go

ord-validator:
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(GOBIN)/ord-validator cmd/validator/main.go
