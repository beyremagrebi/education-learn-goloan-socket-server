export CGO_ENABLED=0

build:
	@go build -o bin/erudaxis main.go

test:
	@go test -v ./...

run: build
	@./bin/erudaxis
