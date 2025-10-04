.PHONY: all build run test clean

all: build

build:
	go build -v -o lowkey main.go

run:
	go run main.go

test:
	go test -v ./...

clean:
	rm -f lowkey
