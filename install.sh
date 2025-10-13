#!/usr/bin/env bash

# Lowkey Installation Script
# Provides options for local or global installation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_color() {
    local color=$1
    shift
    echo -e "${color}$*${NC}"
}

# Check if binary exists
if [[ ! -f "./lowkey" ]]; then
    print_color $YELLOW "Binary not found. Building lowkey..."
    make clean && make build
    if [[ $? -ne 0 ]]; then
        print_color $RED "Build failed. Please check your Go installation."
        exit 1
    fi
    print_color $GREEN "Build successful!"
fi

# Display installation options
print_color $BLUE "Lowkey Installation Options:"
echo "1) Install to ~/.local/bin (local user installation)"
echo "2) Install to /usr/local/bin (system-wide, requires sudo)"
echo "3) Create symlink in ~/.local/bin"
echo "4) Create symlink in /usr/local/bin (requires sudo)"
echo "5) Add current directory to PATH (display instructions)"
echo "6) Cancel installation"
echo ""
read -p "Choose installation method [1-6]: " choice

case $choice in
    1)
        # Local installation to ~/.local/bin
        mkdir -p ~/.local/bin
        cp ./lowkey ~/.local/bin/
        print_color $GREEN "✓ Lowkey installed to ~/.local/bin/"

        # Check if ~/.local/bin is in PATH
        if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
            print_color $YELLOW "\nNote: ~/.local/bin is not in your PATH."
            print_color $YELLOW "Add the following to your shell config (~/.bashrc, ~/.zshrc, etc.):"
            echo 'export PATH="$HOME/.local/bin:$PATH"'
        else
            print_color $GREEN "lowkey is ready to use!"
        fi
        ;;

    2)
        # System-wide installation to /usr/local/bin
        print_color $YELLOW "Installing to /usr/local/bin (requires sudo)..."
        sudo cp ./lowkey /usr/local/bin/
        sudo chmod 755 /usr/local/bin/lowkey
        print_color $GREEN "✓ Lowkey installed to /usr/local/bin/"
        print_color $GREEN "lowkey is ready to use system-wide!"
        ;;

    3)
        # Create symlink in ~/.local/bin
        mkdir -p ~/.local/bin
        ln -sf "$(pwd)/lowkey" ~/.local/bin/lowkey
        print_color $GREEN "✓ Symlink created in ~/.local/bin/"

        # Check if ~/.local/bin is in PATH
        if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
            print_color $YELLOW "\nNote: ~/.local/bin is not in your PATH."
            print_color $YELLOW "Add the following to your shell config (~/.bashrc, ~/.zshrc, etc.):"
            echo 'export PATH="$HOME/.local/bin:$PATH"'
        else
            print_color $GREEN "lowkey is ready to use!"
        fi
        ;;

    4)
        # Create symlink in /usr/local/bin
        print_color $YELLOW "Creating symlink in /usr/local/bin (requires sudo)..."
        sudo ln -sf "$(pwd)/lowkey" /usr/local/bin/lowkey
        print_color $GREEN "✓ Symlink created in /usr/local/bin/"
        print_color $GREEN "lowkey is ready to use system-wide!"
        ;;

    5)
        # Display PATH instructions
        print_color $BLUE "\nTo add the current directory to your PATH:"
        echo ""
        print_color $YELLOW "For bash (~/.bashrc):"
        echo "export PATH=\"$(pwd):\$PATH\""
        echo ""
        print_color $YELLOW "For zsh (~/.zshrc):"
        echo "export PATH=\"$(pwd):\$PATH\""
        echo ""
        print_color $YELLOW "For fish (~/.config/fish/config.fish):"
        echo "set -gx PATH $(pwd) \$PATH"
        echo ""
        print_color $BLUE "After adding the line, reload your shell config:"
        echo "source ~/.bashrc  # or ~/.zshrc, etc."
        ;;

    6)
        print_color $YELLOW "Installation cancelled."
        exit 0
        ;;

    *)
        print_color $RED "Invalid option. Installation cancelled."
        exit 1
        ;;
esac

# Test installation
echo ""
print_color $BLUE "Testing installation..."
if command -v lowkey &> /dev/null; then
    print_color $GREEN "✓ lowkey is accessible from the command line"
    lowkey status || true
else
    print_color $YELLOW "⚠ lowkey is not yet accessible from the current shell."
    print_color $YELLOW "  You may need to reload your shell or update your PATH."
fi