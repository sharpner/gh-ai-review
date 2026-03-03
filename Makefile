BINARY := gh-ai-review

build:
	go build -o $(BINARY) .

test:
	go test ./...

lint:
	golangci-lint run

install: build
	gh extension install .

clean:
	rm -f $(BINARY)

.PHONY: build test lint install clean
