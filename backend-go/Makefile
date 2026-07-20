.PHONY: build test lint run clean

BINARY_NAME=hauk-go
GOFLAGS=-v

build:
	go build $(GOFLAGS) -o $(BINARY_NAME) ./cmd/server/

test:
	go test -race -count=1 ./...

lint:
	go vet ./...
	goimports -l .

run: build
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)
