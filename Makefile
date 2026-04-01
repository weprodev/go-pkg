# Makefile for go-pkg

.PHONY: test audit vet tidy format

# Run all tests with coverage
test:
	go test -v -cover ./...

# Vet code for issues
vet:
	go vet ./...

# Format the code
format:
	go fmt ./...

# Tidy dependencies
tidy:
	go mod tidy

# Run the complete audit suite (formatting, vetting, testing)
audit: tidy format vet test
	@echo "Audit completed successfully ✅"
