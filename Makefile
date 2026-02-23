.PHONY: build install clean

BINARY := tfn
INSTALL_PATH := /usr/local/bin/$(BINARY)

build:
	go build -o $(BINARY) .

install: build
	sudo ln -sf $(CURDIR)/$(BINARY) $(INSTALL_PATH)

clean:
	rm -f $(BINARY)
