# Binary name
BINARY_NAME=email_service

# Go related variables
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_CLEAN=$(GO_CMD) clean
GO_TEST=$(GO_CMD) test
GO_RUN=$(GO_CMD) run
MAIN_PATH=cmd/api/main.go

.PHONY: all build run clean test help

all: build

build:
	$(GO_BUILD) -o bin/$(BINARY_NAME) $(MAIN_PATH)

run:
	$(GO_RUN) $(MAIN_PATH)

test:
	$(GO_TEST) -v ./...

clean:
	$(GO_CLEAN)
	rm -rf bin/

help:
	@echo "Available targets:"
	@echo "  build   : Build the application"
	@echo "  run     : Run the application"
	@echo "  test    : Run tests"
	@echo "  clean   : Remove binary and build artifacts"
	@echo "  help    : Show this help message"
