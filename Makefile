GO := $(or $(shell which go),/usr/local/go/bin/go)
BINARY := _jj
DIST := dist

.PHONY: build clean install test release

build:
	$(GO) build -o $(BINARY) .

install: build
	install -m 755 $(BINARY) /usr/local/bin/$(BINARY)

uninstall:
	rm -f /usr/local/bin/$(BINARY)

clean:
	rm -f $(BINARY)
	rm -rf $(DIST)

test: build
	$(GO) test -v ./...

release:
ifndef VERSION
	$(error VERSION is required. Usage: make release VERSION=v0.1.0)
endif
	mkdir -p $(DIST)/$(VERSION)
	GOOS=darwin  GOARCH=amd64 $(GO) build -o $(DIST)/$(VERSION)/$(BINARY)-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 $(GO) build -o $(DIST)/$(VERSION)/$(BINARY)-darwin-arm64 .
	GOOS=linux   GOARCH=amd64 $(GO) build -o $(DIST)/$(VERSION)/$(BINARY)-linux-amd64 .
	GOOS=linux   GOARCH=arm64 $(GO) build -o $(DIST)/$(VERSION)/$(BINARY)-linux-arm64 .
