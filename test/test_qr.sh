#!/bin/bash
# Test QR code generation and verify URLs are correct
#
# IMPORTANT: This test creates real data by calling the API.
# It backs up and restores your data directory to prevent pollution.

set -e

echo "Testing QR code generation..."

# Backup existing data (test creates real events via API)
BACKUP_DIR="/tmp/sourdough-qr-test-backup-$$"
mkdir -p "$BACKUP_DIR"
if [ -d "data" ]; then
    cp -r data/* "$BACKUP_DIR/" 2>/dev/null || true
fi

# Cleanup function to restore data
cleanup() {
    echo "Restoring original data..."
    rm -rf data/*
    if [ -d "$BACKUP_DIR" ] && [ "$(ls -A $BACKUP_DIR 2>/dev/null)" ]; then
        cp -r "$BACKUP_DIR"/* data/ 2>/dev/null || true
    fi
    rm -rf "$BACKUP_DIR"
}

# Set trap to restore data on exit (even if test fails)
trap cleanup EXIT

# Clean up old QR codes
rm -rf qrcodes_test
mkdir -p qrcodes_test

# Test 1: Reject --help
echo "Test 1: Reject --help flag"
if ./bin/qrgen --help 2>&1 | grep -q "Usage:"; then
    echo "✓ --help shows usage"
else
    echo "✗ --help should show usage"
    exit 1
fi

# Test 2: Reject localhost URLs
echo "Test 2: Reject localhost URLs"
if ./bin/qrgen http://localhost:8080 2>&1 | grep -q "Cannot use localhost"; then
    echo "✓ Localhost URLs rejected"
else
    echo "✗ Should reject localhost URLs"
    exit 1
fi

# Test 3: Generate valid QR codes
echo "Test 3: Generate QR codes with valid URL"
SERVER_URL="http://192.168.1.50:8080"
./bin/qrgen $SERVER_URL > /dev/null 2>&1

# Test 4: Verify QR codes exist
echo "Test 4: Verify QR code files exist"
for file in start.png fed.png mixed.png fold.png qrcodes.pdf; do
    if [ -f "qrcodes/$file" ]; then
        echo "✓ Found qrcodes/$file"
    else
        echo "✗ Missing qrcodes/$file"
        exit 1
    fi
done

# Test 5: Test a QR code URL by making a request
echo "Test 5: Test QR code endpoints work"
if curl -s -X POST http://192.168.1.50:8080/log/mixed | grep -q '"status":"logged"'; then
    echo "✓ QR code endpoint works"
else
    echo "✗ QR code endpoint failed"
    exit 1
fi

echo ""
echo "✓ All QR code tests passed!"
