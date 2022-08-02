SHELL := $(shell which bash)
ENV = /usr/bin/env

.SHELLFLAGS = -c
.ONESHELL: ;

.PHONY: test
test:
	@go test -race -v ./... -cover

.PHONY: lint
lint:
	@go install -mod=readonly github.com/golangci/golangci-lint/cmd/golangci-lint@v1.47.2
	@golangci-lint run
