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
	@$(MAKE) --no-print-directory deps
	@$(MAKE) --no-print-directory telegraf

.PHONY: deps
deps:
	dep ensure -vendor-only

.PHONY: telegraf
telegraf:
	go build -ldflags "$(LDFLAGS)" ./cmd/telegraf


.PHONY: test
test:
	go test -short ./...

.PHONY: fmt
fmt:
	@gofmt -s -w $(filter-out plugins/parsers/influx/machine.go, $(GOFILES))

.PHONY: fmtcheck
fmtcheck:
	@if [ ! -z "$(GOFMT)" ]; then \
		echo "[ERROR] gofmt has found errors in the following files:"  ; \
		echo "$(GOFMT)" ; \
		echo "" ;\
		echo "Run make fmt to fix them." ; \
		exit 1 ;\
	fi

.PHONY: vet
vet:
	@echo 'go vet $$(go list ./... | grep -v ./plugins/parsers/influx)'
	@go vet $$(go list ./... | grep -v ./plugins/parsers/influx) ; if [ $$? -ne 0 ]; then \
		echo ""; \
		echo "go vet has found suspicious constructs. Please remediate any reported errors"; \
		echo "to fix them before submitting code for review."; \
		exit 1; \
	fi

.PHONY: check
check: fmtcheck vet

.PHONY: test-all
test-all: fmtcheck vet
	go test ./...


.PHONY: clean
clean:
	rm -f telegraf
	rm -f telegraf.exe

.PHONY: docker-image
docker-image:
	docker build -f scripts/stretch.docker -t "telegraf:$(COMMIT)" .


.PHONY: static
static:
	@echo "Building static linux binary..."
	@CGO_ENABLED=0 \
	GOOS=linux \
	GOARCH=amd64 \
	go build -ldflags "$(LDFLAGS)" ./cmd/telegraf

