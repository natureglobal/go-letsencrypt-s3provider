VERSION = $(shell godzil show-version)
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X main.revision=$(CURRENT_REVISION) -extldflags \"-static\""
VERBOSE_FLAG = $(if $(VERBOSE),-v)
u := $(if $(update),-u)

export GO111MODULE=on
export CGO_ENABLED=0

.PHONY: deps
deps:
	go get ${u} -d $(VERBOSE_FLAG)
	go mod tidy

.PHONY: devel-deps
devel-deps: deps
	sh -c '\
      tmpdir=$$(mktemp -d); \
      cd $$tmpdir; \
      go get ${u} \
        golang.org/x/lint/golint            \
        github.com/Songmu/godzil/cmd/godzil \
        github.com/tcnksm/ghr;              \
      rm -rf $$tmpdir'

.PHONY: test
test: deps
	go test $(VERBOSE_FLAG) ./...

.PHONY: lint
lint: devel-deps
	golint -set_exit_status ./...

.PHONY: build
build: deps
	go build $(VERBOSE_FLAG) -tags=netgo -installsuffix=netgo -ldflags=$(BUILD_LDFLAGS)

.PHONY: install
install: deps
	go install $(VERBOSE_FLAG) -tags=netgo -installsuffix=netgo -ldflags=$(BUILD_LDFLAGS)

.PHONY: bump
bump: devel-deps
	godzil release

CREDITS: devel-deps go.sum
	godzil credits -w

DIST_DIR = dist/v$(VERSION)
.PHONY: crossbuild
crossbuild: devel-deps
	rm -rf $(DIST_DIR)
	CGO_ENABLED=0 godzil crossbuild -arch=amd64 -os=linux,darwin \
      -build-tags=netgo -build-installsuffix=netgo \
      -build-ldflags=$(BUILD_LDFLAGS) -d $(DIST_DIR) ./cmd/*

.PHONY: upload
upload:
	ghr -body="$$(./godzil changelog --latest -F markdown)" v$(VERSION) $(DIST_DIR)

.PHONY: release
release: bump docker-release

.PHONY: local-release
local-release: bump crossbuild upload

.PHONY: docker-release
docker-release:
	@docker run \
      -v $(PWD):/go-letsencrypt-s3provider \
      -w /go-letsencrypt-s3provider \
      -e GITHUB_TOKEN="$(GITHUB_TOKEN)" \
      --rm \
      golang:1.15.2-alpine3.12 \
      sh -c 'apk add make git && make crossbuild upload'
