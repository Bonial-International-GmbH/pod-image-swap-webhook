.DEFAULT_GOAL := help

TEST_FLAGS ?= -race
PKGS       ?= $(shell go list ./... | grep -v /vendor/)
BINARY     := pod-image-swap-webhook
IMAGE      ?= pod-image-swap-webhook
TAG        ?= latest

.PHONY: all clean

.PHONY: help
help:
	@grep -E '^[a-zA-Z-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "[32m%-12s[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## build pod-image-swap-webhook
	go build \
		-ldflags "-s -w" \
		-o $(BINARY) \
		main.go

.PHONY: docker-build
docker-build: ## build docker image
	docker build -t $(IMAGE):$(TAG) .

.PHONY: test
test: ## run tests
	go test $(TEST_FLAGS) $(PKGS)

.PHONY: vet
vet: ## run go vet
	go vet $(PKGS)

.PHONY: coverage
coverage: ## generate code coverage
	go test $(TEST_FLAGS) -covermode=atomic -coverprofile=coverage.txt $(PKGS)
	go tool cover -func=coverage.txt

.PHONY: lint
lint: ## run golangci-lint
	golangci-lint run
