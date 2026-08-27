BINARY      := chaosbox
BIN_DIR     := bin
PKG         := ./cmd/chaosbox

CONFIG      ?=
FILE        ?=
DATA        ?= data
LOGS        ?= logs
REDIS       ?=
STARTUP_DELAY ?=

RUN_ARGS :=
ifneq ($(CONFIG),)
RUN_ARGS += -config $(CONFIG)
endif
ifneq ($(FILE),)
RUN_ARGS += -file $(FILE)
endif
ifneq ($(REDIS),)
RUN_ARGS += -redis $(REDIS)
endif
ifneq ($(STARTUP_DELAY),)
RUN_ARGS += -startup-delay $(STARTUP_DELAY)
endif

.PHONY: all build run test vet fmt tidy clean docker-build docker-run compose-up compose-down compose-logs help

all: build

## build: compile the chaosbox binary into bin/
build:
	go build -trimpath -o $(BIN_DIR)/$(BINARY) $(PKG)

## run: build and run chaosbox (override via CONFIG=, FILE=, REDIS=, STARTUP_DELAY=)
run: build
	@mkdir -p $(DATA) $(LOGS)
	./$(BIN_DIR)/$(BINARY) $(RUN_ARGS)

## test: run the Go test suite
test:
	go test ./...

## vet: run go vet
vet:
	go vet ./...

## fmt: format all Go source files
fmt:
	gofmt -l -w .

## tidy: tidy go.mod/go.sum
tidy:
	go mod tidy

## clean: remove build artifacts and local data/logs dirs
clean:
	rm -rf $(BIN_DIR) $(DATA) $(LOGS)

## docker-build: build the chaosbox Docker image
docker-build:
	docker build -t $(BINARY) .

## docker-run: build and run the chaosbox Docker image (no config mount required)
docker-run: docker-build
	docker run --rm -p 8080:8080 $(BINARY)

## compose-up: start redis + two peered chaosbox nodes (chaosbox-a, chaosbox-b) via docker compose
compose-up:
	docker compose up --build --pull always -d

## compose-down: stop and remove the docker compose stack (add rm-volumes=1 to also drop volumes)
compose-down:
	docker compose down $(if $(rm-volumes),-v,)

## compose-logs: follow logs from the docker compose stack
compose-logs:
	docker compose logs -f

## help: list available targets
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## /  /'
