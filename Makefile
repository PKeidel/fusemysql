BINARY_NAME=fusemysql
PLATFORMS := linux/amd64
VERSION ?= $(shell git describe --tags --always --long --dirty)
COMMITID = $(shell git rev-parse HEAD)
LDGITHASHFLAG := -X 'main.version=$(VERSION) main.githash=$(COMMITID)'
GOGCCFLAGS := '-s -O3'

export GOAMD64=v3

# split $(PLATFORMS)
# for a list of possible $(PLATFORM) values see: `go tool dist list`
temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

.PHONY: release $(PLATFORMS) dev help test cover bench build build-release build-gcc install run clean update-libs

release: $(PLATFORMS)

$(PLATFORMS):
	GOOS=$(os) GOARCH=$(arch) go build -o '$(BINARY_NAME)-$(VERSION)-$(os)-$(arch)' .

dev: ## Runs the app locally in dev mode
	go run -tags debug .

help: ## Prints all available make commands
	@grep -E '^[a-zA-Z_-]+:' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":"}; {printf "\033[36m%-30s\033[0m\n", $$1}'

test: ## Runs the tests for the app
	go test .

cover: ## Runs the tests with coverage enabled and opens the report in a local browser
	go test -race -coverprofile=coverage.out -v .
	@go tool cover -func coverage.out | tail -n 1 | awk '{ print "Total coverage: " $$3 }'
	@go tool cover -html=coverage.out

bench: ## Runs the benchmarks
	go test -bench=. .

${BINARY_NAME}: *.go
	GOAMD64=v3 go build -o ${BINARY_NAME} -ldflags="$(LDGITHASHFLAG)" .
	@ls -lh ${BINARY_NAME}

build: ${BINARY_NAME}

${BINARY_NAME}-release: *.go
	GOAMD64=v3 go build -o ${BINARY_NAME}-release -ldflags="-s $(LDGITHASHFLAG)" -trimpath .
	@ls -lh ${BINARY_NAME}-release

build-release: ${BINARY_NAME}-release

${BINARY_NAME}-gcc: *.go
	go build -compiler gccgo -gccgoflags="${GOGCCFLAGS}"  -o ${BINARY_NAME}-gcc .
	@ls -lh ${BINARY_NAME}-gcc

build-gcc: ${BINARY_NAME}-gcc

install: ## Installs the app to GOPATH
	go install -ldflags="-s $(LDGITHASHFLAG)" -trimpath .
	@ls -lh $$(which ${BINARY_NAME})

run: build-release ## Starts the prebuild binary
	./${BINARY_NAME}-release

clean: ## Cleans everything
	go clean
	rm ${BINARY_NAME}

update-libs: ## Update all go libs
	go get -u ./...
