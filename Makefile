include .env.dist
export

.SILENT: ;
.ONESHELL: ;
.NOTPARALLEL: ;
.EXPORT_ALL_VARIABLES: ;
Makefile: ;

all: fmt mod build lint test

fmt:
	go fmt ./...

test:
	go test -race -cover ./...

mod:
	go mod tidy
	go mod download

lint:
	~/go/bin/golangci-lint run ./...

help:
	@echo 'Usage: make <TARGETS> ...'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@echo '    help                        Show this help screen'
	@echo '    fmt                         Run gofmt on package sources'
	@echo '    lint                        Run linters'
	@echo '    test                        Run unit tests'
	@echo '    mod                         Update Go modules'
	@echo ''
	@echo 'Targets run by default are: fmt mod build lint test'
	@echo ''
