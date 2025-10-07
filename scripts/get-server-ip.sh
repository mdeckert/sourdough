#!/bin/bash
# Get the local IP address for QR code generation

echo "Finding server IP address..."
echo ""

# Get primary network interface IP (exclude localhost and docker)
IP=$(ip -4 addr show | grep -oP '(?<=inet\s)\d+(\.\d+){3}' | grep -v '^127\.' | grep -v '^172\.17\.' | head -1)

if [ -z "$IP" ]; then
    echo "Error: Could not find local IP address"
    exit 1
fi

echo "Server IP: $IP"
echo "Server URL: http://$IP:8080"
echo ""
echo "Use this URL to generate QR codes:"
echo "  ./bin/qrgen http://$IP:8080"
echo ""
echo "Test server from your phone by visiting:"
echo "  http://$IP:8080/health"
