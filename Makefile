PREFIX := /usr/local
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git rev-parse --short HEAD)
GOFILES ?= $(shell git ls-files '*.go')
GOFMT ?= $(shell gofmt -l -s $(filter-out plugins/parsers/influx/machine.go, $(GOFILES)))
BUILDFLAGS ?=

ifdef GOBIN
PATH := $(GOBIN):$(PATH)
else
PATH := $(subst :,/bin:,$(shell go env GOPATH))/bin:$(PATH)
endif

LDFLAGS := $(LDFLAGS) -X main.commit=$(COMMIT) -X main.branch=$(BRANCH)

.PHONY: all
all:
	@$(MAKE) --no-print-directory telex

.PHONY: telex
telex:
	go mod download
	go build -ldflags "$(LDFLAGS)" ./cmd/telex

telex_linux_x86:
	go mod download
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" ./cmd/telex

.PHONY: fmt
fmt:
	@gofmt -s -w $(filter-out plugins/parsers/influx/machine.go, $(GOFILES))

.PHONY: vet
vet:
	@echo 'go vet $$(go list ./... | grep -v ./plugins/parsers/influx)'
	@go vet $$(go list ./... | grep -v ./plugins/parsers/influx) ; if [ $$? -ne 0 ]; then \
		echo ""; \
		echo "go vet has found suspicious constructs. Please remediate any reported errors"; \
		exit 1; \
	fi

.PHONY: check
check: fmt vet

.PHONY: test
test-all: fmt vet
	go test ./...

testshort:
	go test -shrot ./...

.PHONY: clean
clean:
	rm -f telex

.PHONY: docker-image
docker-image: telex_linux_x86
	docker build -f scripts/smimg.docker -t "telex" .

docker-build:
	docker run -v $(PWD):/work -w /work golang:1.12.4 make 

docker-test:
	docker run -v $(PWD):/work -w /work golang:1.12.4 go test ./...

loctest:
	docker run -i -t -v $(PWD)/etc/telex.conf:/telex.conf -e TELEGRAF_CONFIG_PATH=/telex.conf telex telex --test

docker-alpine: telex_linux_x86
	docker build -f scripts/alpine.docker -t "telexa:$(COMMIT)" .

.PHONY: static
static:
	@echo "Building static linux binary..."
	@CGO_ENABLED=0 \
	GOOS=linux \
	GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS)" ./cmd/telex

