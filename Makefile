# Run the application
run: swag
	@echo "Starting the Rocket State Service..."
	go run main.go

# Run all unit tests
test:
	@echo "Running unit tests..."
	go test -v ./...

# Run tests and print coverage report to console, excluding mock files
coverage: swag
	@echo "Running tests and printing coverage report to console (excluding _mock.go files)..."
	go test -v -coverpkg=./... -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo "Coverage report printed to console."
	@echo "To generate HTML report, run 'go tool cover -html=coverage.out -o coverage.html'"

# Generate Swagger documentation
swag:
	@echo "Generating Swagger documentation..."
	go mod tidy # Ensure all modules are up-to-date
	@$(shell go env GOPATH)/bin/swag init

# Clean up generated files
clean:
	@echo "Cleaning up generated files..."
	rm -f coverage.out coverage.html
	rm -rf docs
	@echo "Cleaned."

