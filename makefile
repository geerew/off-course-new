
## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

# audit: run quality control checks
.PHONY: audit
audit:
	go mod verify
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-ST1003,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	go test -race -buildvcs -vet=off ./...

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## test: run all tests
.PHONY: test
test:
	go test -v ./...

## race_test: run all tests with race detector
.PHONY: race_test
race_test:
	go test -v -race -buildvcs ./...

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

# ==================================================================================== #
# BUILD & RELEASE
# ==================================================================================== #

## local_build: build the application locally
.PHONY: local_build
local_build:
	cd ui && pnpm install && pnpm run build
	go install github.com/goreleaser/goreleaser/v2@latest
	goreleaser build --snapshot --clean

## build: build the application
.PHONY: build
build:
	cd ui && pnpm install && pnpm run build
	go install github.com/goreleaser/goreleaser/v2@latest
	goreleaser build --clean
