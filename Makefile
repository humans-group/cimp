.PHONY: gen
COVERAGE_FILE ?= coverage.txt

#######################################################
######################### GITLAB ######################
#######################################################

prepare:
	echo NOOP

check:
	go vet ./...
	golangci-lint run

gen_modules:
	go mod tidy

gen: gen_modules

test:
	go test -v -race -coverprofile=${COVERAGE_FILE} -covermode=atomic ./...

precompile:
	go build ./...

compile:
ifeq (${BINARY_VERSION},unspecified)
	$(eval BINARY_VERSION:=$(shell ./tag.sh -i))
endif
	go build \
	-ldflags="-s -w \
	-X 'pkg.humans.net/lib/grpc-core/version.Version=${BINARY_VERSION}'" \
	-o bin/binary \
	cmd/${BINARY_NAME}/main.go

new_tag:
	./tag.sh -p

#######################################################
######################### LOCAL #######################
#######################################################
