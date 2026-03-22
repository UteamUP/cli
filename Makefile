BINARY=uteamup
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: build test lint install clean release fmt vet check package-msi package-pkg

build:
	go build $(LDFLAGS) -o bin/$(BINARY) .
	@ln -sf $(BINARY) bin/ut 2>/dev/null || cp bin/$(BINARY) bin/ut
	@echo "Built bin/$(BINARY) and bin/ut"

test:
	go test ./... -v -race -cover

lint:
	golangci-lint run ./...

install:
	go install $(LDFLAGS) .

clean:
	rm -rf bin/ dist/

fmt:
	gofmt -s -w .
	goimports -w . 2>/dev/null || true

vet:
	go vet ./...

check: fmt vet lint test build

release:
	goreleaser release --clean

snapshot:
	goreleaser build --snapshot --clean

package-pkg:
	@echo "Building macOS .pkg installer..."
	./packaging/macos/build-pkg.sh

package-msi:
	@echo "Building Windows MSI installer (requires WiX on Windows/CI)..."
	@echo "Run: wix build packaging/msi/uteamup.wxs -o dist/uteamup.msi"

docs:
	@mkdir -p docs/commands
	go run . docs --dir docs/commands
	@echo "Generated command docs in docs/commands/"
