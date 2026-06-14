.PHONY: proto test cover build lint race

proto:
	protoc --proto_path=proto --go_out=proto --go_opt=paths=source_relative proto/api.proto

build:
	go build ./...

test:
	go test -v -race ./...

cover:
	go test -coverprofile=coverage.out -race ./...
	go tool cover -func=coverage.out | grep total

race:
	go test -v -race ./...

lint:
	golangci-lint run ./...
