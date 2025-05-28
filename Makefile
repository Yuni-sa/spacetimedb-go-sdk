# Default target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  test         - Run all tests (publishes with dbname 'test')"
	@echo "  test-verbose - Run tests with verbose output"
	@echo "  publish      - Publish server with custom dbname (usage: make publish DBNAME=mydb)"
	@echo "  quickstart   - Publish server for quickstart-chat"
	@echo "  help         - Show this help message"

# Publish quickstart-chat modules with configurable database name
.PHONY: publish
publish:
	@if [ -z "$(DBNAME)" ]; then \
		echo "Error: DBNAME is required. Usage: make publish DBNAME=mydb"; \
		exit 1; \
	fi
	spacetime publish $(DBNAME) -s http://localhost:3000 -p ./examples/quickstart-chat/server

# Publish for quickstart-chat
.PHONY: quickstart
quickstart:
	$(MAKE) publish DBNAME=quickstart-chat
	go run ./examples/quickstart-chat/client/main.go

# Run all tests
.PHONY: test
test:
	$(MAKE) publish DBNAME=test
	go test ./... --timeout 30s

# Run tests with verbose output
.PHONY: test-verbose
test-verbose:
	$(MAKE) publish DBNAME=test
	go test -v ./... --timeout 30s

