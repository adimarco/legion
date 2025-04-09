.PHONY: test examples clean

# Default target
all: test examples

# Run all tests with coverage
test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

# Run all examples
examples: examples/concurrent_specialists examples/research_team examples/research_team_simple

examples/concurrent_specialists:
	@echo "\nRunning concurrent specialists example..."
	@cd examples/concurrent_specialists && go run main.go

examples/research_team:
	@echo "\nRunning research team example..."
	@cd examples/research_team && go run main.go

examples/research_team_simple:
	@echo "\nRunning research team simple example..."
	@cd examples/research_team_simple && go run main.go

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@go clean
	@find . -type f -name '*.test' -delete
	@find . -type f -name '*.out' -delete 