.PHONY: all build run test clean

all: build

build:
	go build -v -o lowkey ./cmd/lowkey

run:
	go run ./cmd/lowkey

test:
	go test -v ./...

clean:
	rm -f lowkey
