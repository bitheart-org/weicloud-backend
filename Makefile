APP=weicloud-backend

.PHONY: dev build test tidy

dev:
	go run ./cmd/server

build:
	go build -o bin/$(APP) ./cmd/server

test:
	go test ./...

tidy:
	go mod tidy

