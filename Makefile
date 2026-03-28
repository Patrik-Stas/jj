GO := /usr/local/go/bin/go
BINARY := _jj
VERSION := 0.1.0

.PHONY: build clean install test cross

build:
	$(GO) build -o $(BINARY) .

install: build
	install -m 755 $(BINARY) /usr/local/bin/$(BINARY)

uninstall:
	rm -f /usr/local/bin/$(BINARY)

clean:
	rm -f $(BINARY) $(BINARY)-*

test: build
	$(GO) test -v ./...

# Cross-compile for release
cross:
	GOOS=darwin  GOARCH=amd64 $(GO) build -o $(BINARY)-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 $(GO) build -o $(BINARY)-darwin-arm64 .
	GOOS=linux   GOARCH=amd64 $(GO) build -o $(BINARY)-linux-amd64 .
	GOOS=linux   GOARCH=arm64 $(GO) build -o $(BINARY)-linux-arm64 .
