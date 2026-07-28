# Makefile
BINARY_NAME=create-user
INSTALL_DIR=/usr/local/bin

.PHONY: build install clean deps

deps:
	@if [ ! -f go.mod ]; then go mod init create-user; fi
	go get github.com/charmbracelet/bubbletea@latest
	go get github.com/charmbracelet/bubbles@latest
	go get github.com/charmbracelet/lipgloss@latest
	go mod tidy

build: deps
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) create-user.go

install: build
	@echo "Installing to $(INSTALL_DIR)..."
	sudo cp $(BINARY_NAME) $(INSTALL_DIR)/
	sudo chmod 755 $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "✓ Installed successfully!"
	@echo "Run with: sudo $(BINARY_NAME)"

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -f go.mod go.sum
