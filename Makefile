.PHONY: test
test:
	@go run gotest.tools/gotestsum@latest

.PHONY: fmt
fmt:
	@gofumpt -l -w .
