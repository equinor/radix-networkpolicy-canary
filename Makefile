ENVIRONMENT ?= dev

CONTAINER_REPO ?= radix$(ENVIRONMENT)
DOCKER_REGISTRY	?= $(CONTAINER_REPO).azurecr.io
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
HASH := $(shell git rev-parse HEAD)
TAG := $(BRANCH)-$(HASH)
CURRENT_FOLDER = $(shell pwd)
VERSION		?= ${TAG}
IMAGE_TAG 	?= ${VERSION}
LDFLAGS		+= -s -w

CX_OSES		= linux windows
CX_ARCHS	= amd64

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: lint
lint: bootstrap
	golangci-lint run --max-same-issues 0

.PHONY: build
build:
	docker buildx build -t $(DOCKER_REGISTRY)/radix-networkpolicy-canary:$(VERSION) -t $(DOCKER_REGISTRY)/radix-networkpolicy-canary:$(BRANCH)-$(VERSION) -t $(DOCKER_REGISTRY)/radix-networkpolicy-canary:$(TAG) --platform linux/arm64,linux/amd64 .

.PHONY: deploy
deploy: build
	az acr login --name $(CONTAINER_REPO)
	docker push $(DOCKER_REGISTRY)/radix-networkpolicy-canary:$(VERSION) -t $(DOCKER_REGISTRY)/radix-networkpolicy-canary:$(BRANCH)-$(VERSION) -t $(DOCKER_REGISTRY)/radix-networkpolicy-canary:$(TAG)

.PHONY: verify-generate
verify-generate: tidy
	git diff --exit-code

HAS_GOLANGCI_LINT := $(shell command -v golangci-lint;)

bootstrap:
ifndef HAS_GOLANGCI_LINT
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.3
endif
