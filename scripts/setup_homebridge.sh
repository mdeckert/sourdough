#!/bin/bash
# Setup homebridge for Ecobee temperature integration
# This script installs homebridge and configures it to expose HomeKit accessories via HTTP

set -euo pipefail

echo "========================================="
echo "Homebridge Setup for Ecobee Integration"
echo "========================================="
echo ""

# Check if running as root
if [ "$EUID" -eq 0 ]; then
    echo "Please run as regular user (not root)"
    exit 1
fi

# Check for Node.js
if ! command -v node &> /dev/null; then
    echo "Node.js not found. Installing..."
    curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
    sudo apt-get install -y nodejs
else
    echo "✓ Node.js found: $(node --version)"
fi

# Check for npm
if ! command -v npm &> /dev/null; then
    echo "npm not found. Installing..."
    sudo apt-get install -y npm
else
    echo "✓ npm found: $(npm --version)"
fi

echo ""
echo "Installing homebridge..."
sudo npm install -g --unsafe-perm homebridge homebridge-config-ui-x

echo ""
echo "Installing homebridge HTTP API plugin..."
sudo npm install -g homebridge-http-webhooks

echo ""
echo "Creating homebridge directory..."
mkdir -p ~/.homebridge

echo ""
echo "Creating initial config..."
cat > ~/.homebridge/config.json << 'EOF'
{
    "bridge": {
        "name": "Homebridge",
        "username": "CC:22:3D:E3:CE:30",
        "port": 51826,
        "pin": "031-45-154"
    },
    "accessories": [],
    "platforms": [
        {
            "name": "Config",
            "port": 8581,
            "platform": "config"
        },
        {
            "platform": "HttpWebhooks",
            "webhook_port": "51828",
            "webhook_listen_host": "0.0.0.0"
        }
    ]
}
EOF

echo ""
echo "Creating systemd service..."
sudo tee /etc/systemd/system/homebridge.service > /dev/null << EOF
[Unit]
Description=Homebridge
After=network-online.target

[Service]
Type=simple
User=$USER
ExecStart=/usr/bin/homebridge -I
Restart=on-failure
RestartSec=10
KillMode=process

[Install]
WantedBy=multi-user.target
EOF

echo ""
echo "Enabling and starting homebridge..."
sudo systemctl daemon-reload
sudo systemctl enable homebridge
sudo systemctl start homebridge

echo ""
echo "========================================="
echo "Homebridge Installation Complete!"
echo "========================================="
echo ""
echo "Next steps:"
echo "  1. Open homebridge UI: http://192.168.1.50:8581"
echo "  2. Login with default credentials (will prompt to change)"
echo "  3. Pair with HomeKit:"
echo "     - Open Home app on iPhone"
echo "     - Add Accessory"
echo "     - Scan QR code or enter PIN: 031-45-154"
echo "  4. Once paired, your Ecobee should appear in homebridge"
echo "  5. Note the exact accessory name for configuration"
echo ""
echo "Status: sudo systemctl status homebridge"
echo "Logs:   sudo journalctl -u homebridge -f"
echo ""
