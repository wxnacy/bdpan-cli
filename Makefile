
.PHONY: install install-completion

# Get the GOPATH
GOPATH:=$(shell go env GOPATH)

install:
	@echo "Installing bdpan to $(GOPATH)/bin..."
	@go build -ldflags "-X main.Version=dev" -o $(GOPATH)/bin/bdpan .
	@echo "bdpan installed successfully."
	@$(MAKE) install-completion

install-completion:
	@echo "Generating completion scripts..."
	@go run ./scripts/completion/main.go
	@echo "Installing completion scripts..."
	@mkdir -p ~/.zsh/completions
	@cp ./scripts/completion/bdpan.zsh ~/.zsh/completions/_bdpan
	@echo "Completion scripts installed. Please restart your shell."

