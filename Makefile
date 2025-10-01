
APPS := openai gemini
APP ?= openai

include ./examples/$(APP)/.env
export

example: build-example run-example

lint: tidy
	@echo "Running linters..."
	@gofmt -s -w .
	@golangci-lint run --fix

tidy:
	@echo "Tidying up Go modules..."
	@go mod tidy

.PHONY: example lint tidy

build-example:
	@echo "Building example: $(APP)"
	@go build -o bin/$(APP) ./examples/$(APP)/main.go

run-example: build-example
	@echo "Running example: $(APP)\n"
	@./bin/$(APP)

clean:
	@echo "Cleaning up..."
	@rm -f bin/*
	@echo "Cleanup complete."
