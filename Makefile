default: build

# Get the current version from git tags
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")

# Build the provider
.PHONY: build
build:
	go build -o terraform-provider-warpgate

# Install the provider locally for testing
.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/Thunderbottom/warpgate/$(VERSION)/$(shell go env GOOS)_$(shell go env GOARCH)
	cp terraform-provider-warpgate ~/.terraform.d/plugins/registry.terraform.io/Thunderbottom/warpgate/$(VERSION)/$(shell go env GOOS)_$(shell go env GOARCH)/

# Generate documentation
.PHONY: docs
docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

# Format Go code
.PHONY: fmt
fmt:
	go fmt ./...

# Lint the code
.PHONY: lint
lint:
	golangci-lint run ./...

# Clean build artifacts
.PHONY: clean
clean:
	rm -f terraform-provider-warpgate
