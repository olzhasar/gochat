run:
	go run .

build:
	go build -o gochat .

test:
	go test -v ./...

deploy:
	flyctl deploy

.PHONY: run build test deploy
