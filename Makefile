.PHONY: gen
COVERAGE_FILE ?= coverage.txt

check:
	go vet ./...
	golangci-lint run

gen_modules:
	go mod tidy

gen: gen_modules

precompile:
	go build ./...
