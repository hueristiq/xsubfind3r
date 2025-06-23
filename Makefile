SHELL = /bin/sh

# --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
# --- Setup ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
# --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

.PHONY: install-lefthook install-golangci-lint

install-lefthook:
	(command -v lefthook || go install github.com/evilmartians/lefthook@latest) && lefthook install

install-golangci-lint:
	command -v golangci-lint || go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.0.2

# --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
# --- Go (Golang) ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------
# --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

.PHONY: go-mod-clean go-mod-tidy go-mod-update go-fmt go-lint go-test go-build go-install

go-mod-clean:
	go clean -modcache

go-mod-tidy:
	go mod tidy

go-mod-update:
	go get -f -t -u ./...
	go get -f -u ./...

go-fmt: install-golangci-lint
	golangci-lint fmt ./...

go-lint: go-fmt
	golangci-lint run ./...

go-test:
	go test -v -race ./...

go-build:
	go build -v -ldflags '-s -w' -o bin/xsubfind3r cmd/xsubfind3r/main.go

go-install:
	go install -v ./...

# --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
# --- Docker ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
# --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

DOCKERFILE := ./Dockerfile

IMAGE_NAME = hueristiq/xsubfind3r
IMAGE_TAG = $(shell cat internal/configuration/configuration.go | grep "VERSION =" | sed 's/.*VERSION = "\([0-9.]*\)".*/\1/')
IMAGE = $(IMAGE_NAME):$(IMAGE_TAG)

.PHONY: docker-build

docker-build:
	docker build -f $(DOCKERFILE) -t $(IMAGE) -t $(IMAGE_NAME):latest .

# --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
# --- Help -----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
# --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

.PHONY: help

help:
	@echo ""
	@echo "Available commands:"
	@echo ""
	@echo " Setup:"
	@echo ""
	@echo "  install-lefthook .............. Install lefthook."
	@echo "  install-golangci-lint ......... Install golangci-lint."
	@echo ""
	@echo " Go (Golang):"
	@echo ""
	@echo "  go-mod-clean ............. Clean Go module cache."
	@echo "  go-mod-tidy .............. Tidy Go modules."
	@echo "  go-mod-update ............ Update Go modules."
	@echo "  go-fmt ................... Format Go code."
	@echo "  go-lint .................. Lint Go code."
	@echo "  go-test .................. Run Go tests."
	@echo "  go-build ................. Build Go program."
	@echo "  go-install ............... Install Go program."
	@echo ""
	@echo " Docker:"
	@echo ""
	@echo "  docker-build ............. Build Docker image."
	@echo ""
	@echo " Help:"
	@echo ""
	@echo "  help ..................... Display this help information."
	@echo ""

.DEFAULT_GOAL = help