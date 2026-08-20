BINARY_NAME=explain
BUILD_DIR=bin

ifeq ($(shell id -u), 0)
    PREFIX?=/usr/local
else
    PREFIX?=$(HOME)/.local
endif

.PHONY: all build test clean install uninstall

all: test build

build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/explain

test:
	go test -v ./tests/...

install: build
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 755 $(BUILD_DIR)/$(BINARY_NAME) $(DESTDIR)$(PREFIX)/bin/$(BINARY_NAME)

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/$(BINARY_NAME)

clean:
	rm -rf $(BUILD_DIR)
