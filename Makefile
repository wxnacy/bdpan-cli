
.PHONY: install

GO ?= go

# Get the GOPATH
GOPATH:=$(shell go env GOPATH)

install:
	@echo "Installing bdpan to $(GOPATH)/bin..."
	$(GO) install ./cmd/bdpan
	@echo "bdpan installed successfully."
	@bdpan completion zsh > ~/.zsh/completions/_bdpan
	@echo "Completion scripts installed. Please restart your shell."
