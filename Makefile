export CGO_ENABLED=0

build:
	@go build -o bin/square main.go

test:
	@go test -v ./...

run: build
	@./bin/erudaxis
