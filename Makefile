build:
	@go build -o bin/square main.go

test:
	@go test -v ./...

run: build
	@./bin/square
