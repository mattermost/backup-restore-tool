## Docker Build Versions
DOCKER_BUILDER_SERVER_IMAGE = golang:1.19
DOCKER_BASE_IMAGE = alpine:3.17

################################################################################

export GOBIN ?= $(PWD)/bin
GO ?= $(shell command -v go 2> /dev/null)
GOFLAGS ?= $(GOFLAGS:)
IMAGE ?= mattermost/backup-restore-tool:test

TRIVY_SEVERITY := CRITICAL
TRIVY_EXIT_CODE := 1
TRIVY_VULN_TYPE := os,library

export GO111MODULE=on

all: check-style ## Checks the code style, tests, builds and bundles.

.PHONY: install
install: # Installs Backup Restore Tool on local machine.
	go install ./cmd/backup-restore-tool

.PHONY: unittest
unittest: # Run unit tests.
	go test ./... -v -covermode=count -coverprofile=coverage.out

.PHONY: build-image
build-image: ## Build the docker image of Backup Restore Tool
	@echo Building Container Image
	docker build \
	--build-arg DOCKER_BUILDER_SERVER_IMAGE=$(DOCKER_BUILDER_SERVER_IMAGE) \
	--build-arg DOCKER_BASE_IMAGE=$(DOCKER_BASE_IMAGE) \
	. -t $(IMAGE) \
	--no-cache

.PHONY: check-style
check-style: govet lint ## Runs govet and gofmt against all packages.
	@echo Checking for style guide compliance
	$(GO) fmt ./...

.PHONY: lint
lint: ## Runs lint against all packages.
	@echo Running lint
	env GO111MODULE=off $(GO) get -u golang.org/x/lint/golint
	$(GOBIN)/golint -set_exit_status $(./... | grep -v /blapi/)
	@echo lint success

.PHONY: vet
govet: ## Runs govet against all packages.
	@echo Running govet
	$(GO) vet ./...
	@echo Govet success

.PHONY: build
build: ## Build go binary.
	@echo Building binary
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags '$(LDFLAGS)' -gcflags all=-trimpath=$(PWD) -asmflags all=-trimpath=$(PWD) -a -installsuffix cgo -o build/_output/bin/backup-restore-tool  ./cmd/backup-restore-tool

.PHONY: e2e
e2e: ## Run e2e test.
	@echo Installing Backup Restore Tool
	go install ./cmd/backup-restore-tool
	go test -tags=e2e -v ./tests

.PHONY: e2e-s3-cleanup
e2e-s3-cleanup: ## Removes backup file created in Amazon S3 by e2e test.
	aws s3 rm s3://${BRT_STORAGE_BUCKET}/backup-restore-e2e-test-key

.PHONY: build-image
scan: build-image
	@echo running trivy
	@trivy image --format table --exit-code $(TRIVY_EXIT_CODE) --ignore-unfixed --vuln-type $(TRIVY_VULN_TYPE) --severity $(TRIVY_SEVERITY) $(IMAGE)

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' ./Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
