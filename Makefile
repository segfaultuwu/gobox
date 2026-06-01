APP := gobox

PREFIX ?= /usr/local
DESTDIR ?=
BINDIR := $(DESTDIR)$(PREFIX)/bin

GO ?= go
GOFLAGS ?=
LDFLAGS ?=

BUILD_DIR := build
BIN := $(BUILD_DIR)/$(APP)

<<<<<<< HEAD
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
	mount \
	umount \
	dmesg \
	ps \
	help

=======
>>>>>>> b7b7a7e63e6012585a7892e22886fbc2d37f09ca
.PHONY: all build debug release run clean distclean install uninstall links unlink \
	fmt vet test check mod tidy list-applets doctor

all: build

build:
	@mkdir -p $(BUILD_DIR)
<<<<<<< HEAD
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/gobox
=======
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN) .
>>>>>>> b7b7a7e63e6012585a7892e22886fbc2d37f09ca

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

<<<<<<< HEAD
links:
	@install -d "$(BINDIR)"
	@for applet in $(APPLETS); do \
		if [ "$$applet" = "$(APP)" ]; then \
			continue; \
		fi; \
=======
links: build
	@install -d "$(BINDIR)"
	@for applet in $$("./$(BIN)" applets); do \
		if [ "$$applet" = "$(APP)" ]; then \
			continue; \
		fi; \
		if [ "$$applet" = "applets" ]; then \
			continue; \
		fi; \
>>>>>>> b7b7a7e63e6012585a7892e22886fbc2d37f09ca
		ln -sfn "$(APP)" "$(BINDIR)/$$applet"; \
		echo "linked $(BINDIR)/$$applet -> $(APP)"; \
	done

uninstall: unlink
	rm -f "$(BINDIR)/$(APP)"
	@echo "removed $(BINDIR)/$(APP)"

<<<<<<< HEAD
unlink:
	@for applet in $(APPLETS); do \
=======
unlink: build
	@for applet in $$("./$(BIN)" applets); do \
>>>>>>> b7b7a7e63e6012585a7892e22886fbc2d37f09ca
		if [ -L "$(BINDIR)/$$applet" ]; then \
			rm -f "$(BINDIR)/$$applet"; \
			echo "removed $(BINDIR)/$$applet"; \
		fi; \
	done

<<<<<<< HEAD
list-applets:
	@for applet in $(APPLETS); do \
		echo "$$applet"; \
	done

doctor:
=======
list-applets: build
	@./$(BIN) applets

doctor: build
>>>>>>> b7b7a7e63e6012585a7892e22886fbc2d37f09ca
	@echo "app:       $(APP)"
	@echo "prefix:    $(PREFIX)"
	@echo "destdir:   $(DESTDIR)"
	@echo "bindir:    $(BINDIR)"
	@echo "go:        $$($(GO) version)"
<<<<<<< HEAD
	@echo "applets:   $(words $(APPLETS))"
=======
	@echo "applets:"
	@./$(BIN) applets | sed 's/^/  /'
>>>>>>> b7b7a7e63e6012585a7892e22886fbc2d37f09ca
