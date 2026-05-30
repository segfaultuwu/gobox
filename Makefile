APP := gobox
PREFIX ?= /usr/local
BINDIR := $(PREFIX)/bin

GO := go
GOFLAGS :=

APPLETS := \
	echo \
	cat \
	pwd \
	ls \
	mkdir \
	rm \
	touch \
	cp \
	mv \
	head \
	tail \
	whoami \
	uname \
	clear \
	true \
	false \
	yes \
	sleep

.PHONY: all build run clean install uninstall links unlink fmt vet test

all: build

build:
	$(GO) build $(GOFLAGS) -o $(APP) ./cmd/gobox

run: build
	./$(APP)

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

clean:
	rm -f $(APP)

install: build
	install -Dm755 $(APP) $(BINDIR)/$(APP)
	$(MAKE) links PREFIX=$(PREFIX)

links:
	@for applet in $(APPLETS); do \
		ln -sf $(BINDIR)/$(APP) $(BINDIR)/$$applet; \
		echo "linked $$applet -> $(APP)"; \
	done

uninstall: unlink
	rm -f $(BINDIR)/$(APP)

unlink:
	@for applet in $(APPLETS); do \
		if [ -L "$(BINDIR)/$$applet" ]; then \
			rm -f "$(BINDIR)/$$applet"; \
			echo "removed $(BINDIR)/$$applet"; \
		fi; \
	done
