# Makefile - Chanify server

PACKAGE?=$(shell go list -m)
PROJECT_NAME?=$(notdir $(PACKAGE))

## Env variables
VERSION?=$(shell sed -n 's/^[[:space:]]*Version = "\([^"]*\)"/\1/p' cmd/version.go)
ifeq ($(GITHUB_SHA),)
	GITHUB_SHA=$(shell git rev-parse HEAD)
endif
GIT_COMMIT=$(shell echo ${GITHUB_SHA}|cut -c1-7)
BUILD_TIME=$(shell date -u +%FT%TZ)
LDFLAGS= -ldflags "\
	-X ${PACKAGE}/cmd.GitCommit=${GIT_COMMIT} \
	-X ${PACKAGE}/cmd.BuildTime=${BUILD_TIME} \
	-X ${PACKAGE}/cmd.Version=${VERSION}"

# Command
lint:
	@echo Lint ${PACKAGE}
	@golangci-lint run ./...

# TEST_FLAGS can be overridden for environments without race detector support.
TEST_FLAGS?=-race

test:
	@echo Test ${PACKAGE}
	@go test $(TEST_FLAGS) ./cmd
	@go test $(TEST_FLAGS) -tags test -covermode atomic -coverprofile coverage.raw.out -coverpkg ./... ./...
	@awk '!/\.pb\.go:/' coverage.raw.out > coverage.out
	@rm coverage.raw.out
	@go tool cover -func coverage.out

cover: test
	@go tool cover -html coverage.out

build:
	@echo Build ${PACKAGE}
	@go build ${LDFLAGS} ${PACKAGE}

install:
	@echo Build ${PACKAGE}
	@go install ${LDFLAGS} ${PACKAGE}

.PHONY: lint test cover build install
