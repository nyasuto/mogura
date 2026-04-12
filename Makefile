BINARY := mogura
GOFILES := $(shell find . -name '*.go' -not -path './vendor/*')

.PHONY: all build test vet fmt lint quality clean install

all: quality build

build:
	go build -o $(BINARY) .

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l $(GOFILES)
	@if [ -n "$$(gofmt -l $(GOFILES))" ]; then \
		echo "gofmt found unformatted files"; exit 1; \
	fi

quality: vet fmt test

clean:
	rm -f $(BINARY)

install:
	go install .
