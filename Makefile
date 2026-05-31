APP := gobox

PREFIX ?= /usr/local
DESTDIR ?=
BINDIR := $(DESTDIR)$(PREFIX)/bin

GO ?= go
GOFLAGS ?=
LDFLAGS ?=

BUILD_DIR := build
BIN := $(BUILD_DIR)/$(APP)

APPLETS := \
	sh \
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
	sleep \
	which \
	env \
	export \
	unset \
	cd \
	exit \
	init \
	fetch \
	help

.PHONY: all build debug release run clean distclean install uninstall links unlink \
	fmt vet test check mod tidy list-applets doctor

all: build

build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/gobox

debug: GOFLAGS += -gcflags="all=-N -l"
debug: build

release: LDFLAGS += -s -w
release: build

run: build
	./$(BIN) sh

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

check: fmt vet test

mod:
	$(GO) mod download

tidy:
	$(GO) mod tidy

clean:
	rm -rf $(BUILD_DIR)

distclean: clean
	rm -f $(APP)

install: release
	install -d "$(BINDIR)"
	install -m 755 "$(BIN)" "$(BINDIR)/$(APP)"
	$(MAKE) links PREFIX="$(PREFIX)" DESTDIR="$(DESTDIR)"

links:
	@install -d "$(BINDIR)"
	@for applet in $(APPLETS); do \
		if [ "$$applet" = "$(APP)" ]; then \
			continue; \
		fi; \
		ln -sfn "$(APP)" "$(BINDIR)/$$applet"; \
		echo "linked $(BINDIR)/$$applet -> $(APP)"; \
	done

uninstall: unlink
	rm -f "$(BINDIR)/$(APP)"
	@echo "removed $(BINDIR)/$(APP)"

unlink:
	@for applet in $(APPLETS); do \
		if [ -L "$(BINDIR)/$$applet" ]; then \
			rm -f "$(BINDIR)/$$applet"; \
			echo "removed $(BINDIR)/$$applet"; \
		fi; \
	done

list-applets:
	@for applet in $(APPLETS); do \
		echo "$$applet"; \
	done

doctor:
	@echo "app:       $(APP)"
	@echo "prefix:    $(PREFIX)"
	@echo "destdir:   $(DESTDIR)"
	@echo "bindir:    $(BINDIR)"
	@echo "go:        $$($(GO) version)"
	@echo "applets:   $(words $(APPLETS))"
