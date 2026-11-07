.PHONY: dev build test fmt vet run

AIR_BIN ?= $(shell go env GOPATH)/bin/air

dev:
	@if [ -x "$(AIR_BIN)" ]; then \
		APP_ENV=development "$(AIR_BIN)" -c .air.toml; \
	else \
		echo "Air Go belum terpasang di $(AIR_BIN); menjalankan go run tanpa hot reload"; \
		APP_ENV=development go run ./cmd/server; \
	fi

run:
	go run ./cmd/server

build:
	go build -buildvcs=false -trimpath -ldflags="-s -w" -o bin/gomvc ./cmd/server

test:
	go test ./...

fmt:
	gofmt -w cmd internal

vet:
	go vet ./...
