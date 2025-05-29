.PHONY: build run test clean deps help

# Build the application
build:
	go build -o bin/premier-league-simulator .

# Run the application
run: build
	./bin/premier-league-simulator

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f premier_league.db

# Install dependencies
deps:
	go mod download
	go mod tidy

# Database commands
reset-db:
	rm -f premier_league.db

# Help
help:
	@echo "Available commands:"
	@echo "  build    - Build the application"
	@echo "  run      - Run the desktop GUI application"
	@echo "  test     - Run tests"
	@echo "  clean    - Clean build artifacts"
	@echo "  deps     - Install dependencies"
	@echo "  reset-db - Reset database"
	@echo "  help     - Show this help message" 