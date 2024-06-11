MAIN_PACKAGE_PATH := .
BINARY_NAME := off-course
WINDOWS_BINARY_NAME := ${BINARY_NAME}_windows_amd64.exe
LINUX_BINARY_NAME := ${BINARY_NAME}_linux_amd64
DARWIN_BINARY_NAME := ${BINARY_NAME}_darwin_amd64


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
	go test -v -race -buildvcs ./...

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

## build_ui: build the UI
.PHONY: build_ui
build_ui:
	cd ui && pnpm install && pnpm run build

## build: build the application
.PHONY: build
build: build_ui
	env GOOS=windows GOARCH=amd64 go build -o=dist/${WINDOWS_BINARY_NAME} ${MAIN_PACKAGE_PATH}
	env GOOS=linux GOARCH=amd64 go build -o=dist/${LINUX_BINARY_NAME} ${MAIN_PACKAGE_PATH}	
	env GOOS=darwin GOARCH=amd64 go build -o=dist/${DARWIN_BINARY_NAME} ${MAIN_PACKAGE_PATH}
## dev: run in dev
dev:
	go run main.go

## run: build and run the application
.PHONY: run
run: build
	dist/${BINARY_NAME}

## run/live: run the application with reloading on file changes
.PHONY: run/live
run/live:
	go run github.com/cosmtrek/air@v1.43.0 \
	--build.cmd "make build" --build.bin "dist/${BINARY_NAME}" --build.delay "100" \
	--build.exclude_dir "" \
	--build.include_ext "go, tpl, tmpl, html, css, scss, js, ts, sql, jpeg, jpg, gif, png, bmp, svg, webp, ico" \
	--misc.clean_on_exit "true"