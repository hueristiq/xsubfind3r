SHELL = /bin/sh

# --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
# --- Setup ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
# --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

.PHONY: install-lefthook

install-lefthook:
	(command -v lefthook || go install github.com/evilmartians/lefthook@latest) && lefthook install

# --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
# --- Go(Golang) -----------------------------------------------------------------------------------------------------------------------------------------------------------------------------
# --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------

.PHONY: go-mod-clean
go-mod-clean:
	go clean -modcache

.PHONY: go-mod-tidy
go-mod-tidy:
	go mod tidy

.PHONY: go-mod-update
go-mod-update:
	go get -f -t -u ./...
	go get -f -u ./...

.PHONY: go-fmt
go-fmt:
	(command -v golangci-lint || go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.0.2) && golangci-lint fmt ./...

.PHONY: go-lint
go-lint: go-fmt
	(command -v golangci-lint || go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.0.2) && golangci-lint run ./...

.PHONY: go-test
go-test:
	go test -v -race ./...

.PHONY: go-build
go-build:
	go build -v -ldflags '-s -w' -o bin/xsubfind3r cmd/xsubfind3r/main.go

.PHONY: go-install
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
	@echo "  install-lefthook ......... Install lefthook (Git hooks manager)."
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