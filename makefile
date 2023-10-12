DISTROS="linux/amd64 darwin/amd64 windows/amd64"
PACKAGES := $(shell go list ./... | grep -v /ui | grep -v /migrations)
VERSION := $(shell cat VERSION)

.PHONY: test
test:
	@go clean -testcache
	@echo "mode: count" > coverage-all.out
	@$(foreach pkg,$(PACKAGES),  \
		go test -p=1 -cover -covermode=count -coverprofile=coverage.out ${pkg}; \
		tail -n +2 coverage.out >> coverage-all.out;)

.PHONY: test-cover
test-cover: test
	go tool cover -html=coverage-all.out -o coverage.html