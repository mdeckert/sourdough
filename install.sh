#!/bin/bash
set -e

echo "====================================="
echo "Sourdough Logger Installation Script"
echo "====================================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed"
    echo ""
    echo "Please install Go first:"
    echo "  wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz"
    echo "  sudo rm -rf /usr/local/go"
    echo "  sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz"
    echo "  echo 'export PATH=\$PATH:/usr/local/go/bin' >> ~/.bashrc"
    echo "  source ~/.bashrc"
    exit 1
fi

echo "✓ Go is installed: $(go version)"
echo ""

# Download dependencies
echo "Downloading dependencies..."
go mod download
go mod tidy
echo "✓ Dependencies downloaded"
echo ""

# Build binaries
echo "Building server..."
make server
echo "✓ Server built"
echo ""

echo "Building CLI..."
make cli
echo "✓ CLI built"
echo ""

echo "Building QR generator..."
make qrgen
echo "✓ QR generator built"
echo ""

# Prompt for installation options
echo "Installation Options:"
echo "  1. Test locally (no service installation)"
echo "  2. Install as systemd service (recommended)"
echo ""
read -p "Choose option (1 or 2): " option

if [ "$option" = "1" ]; then
    echo ""
    echo "✓ Build complete! Test with:"
    echo "  ./bin/sourdough-server    # Start server"
    echo "  ./bin/sourdough start     # Start a bake"
    echo "  ./bin/sourdough status    # Check status"
    echo ""

elif [ "$option" = "2" ]; then
    echo ""
    echo "Installing systemd service..."
    make install-service
    echo "✓ Service installed and started"
    echo ""

    echo "Installing CLI to /usr/local/bin..."
    make install
    echo "✓ CLI installed"
    echo ""

    echo "Service status:"
    sudo systemctl status sourdough --no-pager -l
    echo ""

else
    echo "Invalid option"
    exit 1
fi

# Generate QR codes
echo ""
read -p "Generate QR codes now? (y/n): " gen_qr

if [ "$gen_qr" = "y" ]; then
    echo ""
    echo "Finding server IP..."
    ./scripts/get-server-ip.sh
    echo ""

    read -p "Enter server URL (e.g., http://192.168.1.100:8080): " server_url

    if [ -n "$server_url" ]; then
        ./bin/qrgen "$server_url"
        echo ""
        echo "✓ QR codes generated in ./qrcodes/"
        echo "  Print qrcodes/sheet.png and stick on fridge!"
    fi
fi

echo ""
echo "====================================="
echo "Installation Complete!"
echo "====================================="
echo ""
echo "Next steps:"
echo "  1. Start a bake: sourdough start"
echo "  2. Log events: sourdough log mixed"
echo "  3. Check status: sourdough status"
echo "  4. Print QR codes from ./qrcodes/"
echo ""
echo "Documentation:"
echo "  - README.md: Overview and usage"
echo "  - SETUP.md: Detailed setup instructions"
echo "  - QUICK_START.md: Daily workflow guide"
echo ""
echo "Happy baking! 🍞"
