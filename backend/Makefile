run:
	go run ./cmd/

build:
	go build -o gochat ./cmd/

test:
	go test -v ./...

test_race:
	go test -v ./... --race

test_watch:
	rg --files -g '*.go' | entr -cn $(MAKE) test

.PHONY: run build test test_watch
