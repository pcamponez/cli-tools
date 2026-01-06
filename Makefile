# CLI Tools Makefile

# Get all command directories
CMDS := $(notdir $(wildcard cmd/*))
BINDIR := ./bin
INSTALLDIR := $(HOME)/go/bin

.PHONY: all build install clean help $(CMDS)

# Default target
all: build

# Build all commands
build:
	@echo "Building all commands..."
	@mkdir -p $(BINDIR)
	@for cmd in $(CMDS); do \
		echo "  Building $$cmd..."; \
		go build -o $(BINDIR)/$$cmd ./cmd/$$cmd; \
	done
	@echo "Done! Binaries are in $(BINDIR)/"

# Install all commands to ~/go/bin
install: build
	@echo "Installing to $(INSTALLDIR)..."
	@mkdir -p $(INSTALLDIR)
	@for cmd in $(CMDS); do \
		cp $(BINDIR)/$$cmd $(INSTALLDIR)/; \
	done
	@echo "Done! Commands installed to $(INSTALLDIR)/"
	@echo ""
	@echo "Make sure $(INSTALLDIR) is in your PATH:"
	@echo '  export PATH="$$HOME/go/bin:$$PATH"'

# Build individual commands
$(CMDS):
	@mkdir -p $(BINDIR)
	go build -o $(BINDIR)/$@ ./cmd/$@

# Clean built binaries
clean:
	@echo "Cleaning..."
	@rm -rf $(BINDIR)
	@echo "Done!"

# Uninstall commands from ~/go/bin
uninstall:
	@echo "Uninstalling from $(INSTALLDIR)..."
	@for cmd in $(CMDS); do \
		rm -f $(INSTALLDIR)/$$cmd; \
	done
	@echo "Done!"

# Show help
help:
	@echo "Available targets:"
	@echo "  make build     - Build all commands to ./bin/"
	@echo "  make install   - Build and install to ~/go/bin/"
	@echo "  make clean     - Remove built binaries"
	@echo "  make uninstall - Remove installed commands"
	@echo "  make <cmd>     - Build a specific command"
	@echo ""
	@echo "Available commands:"
	@for cmd in $(CMDS); do echo "  $$cmd"; done
