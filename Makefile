BINARY := bin/razer-pair
VERSION ?= dev

.PHONY: build test vet check clean

build:
	go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o $(BINARY) ./cmd/razer-pair

test:
	go test ./...

vet:
	go vet ./...

check: test vet

clean:
	go clean
	rm -f $(BINARY)
