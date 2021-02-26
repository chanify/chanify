# Makefile - Chanify server

PACKAGE?=$(shell go list .|head -1)
PROJECT_NAME?=$(notdir $(PACKAGE))

## Env variables
TAG_COMMIT=$(shell git rev-list --tags=v* --max-count=1)
COMMIT_REF_NAME=$(shell git rev-parse --abbrev-ref HEAD)
ifeq ($(TAG_COMMIT),)
	MAIN_VER=0.0.0
else
	MAIN_VER=$(shell git describe --tags ${TAG_COMMIT}|cut -c2-)
endif
ifeq ($(GITHUB_SHA),)
	GITHUB_SHA=$(shell git rev-parse HEAD)
endif
ifeq ($(COMMIT_REF_NAME), main)
	TAGS=latest ${MAIN_VER}
else
	TAGS=${COMMIT_REF_NAME}
endif


VERSION=${MAIN_VER}
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

test:
	@echo Test ${PACKAGE}
	@go list -f '{{if gt (len .TestGoFiles) 0}}"go test -tags test -covermode count -coverprofile {{.Name}}.coverprofile -coverpkg ./... {{.ImportPath}}"{{end}}' ./... | xargs -I {} sh -c {}
	@gocovmerge `ls *.coverprofile` | grep -v ".pb.go" > coverage.out
	@go tool cover -func coverage.out | grep total

cover: test
	@go tool cover -html coverage.out

build:
	@echo Build ${PACKAGE}
	@go build ${LDFLAGS} ${PACKAGE}

.PHONY: lint test cover build
