.PHONY: clean mocks test test-verbose

# Clean build artifacts and generated files
clean:
	rm -rf mocks/
	find . -type f -name '*.test' -delete
	find . -type f -name 'coverage.out' -delete

# Generate mocks using mockery
mocks:
	go generate ./...

# Run tests
test:
	go test -race ./...

# Run tests with verbose output
test-verbose:
	go test -v -race ./...

# Run tests with coverage
test-coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Default target
all: clean mocks test 