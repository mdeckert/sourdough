#!/bin/bash
# Test image upload and retrieval functionality
# This test creates a temporary image, uploads it, and verifies it can be retrieved

set -euo pipefail

BASE_URL="http://192.168.1.50:8080"
DATA_DIR="./data"
BACKUP_DIR=$(mktemp -d)

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# Backup and restore functions
backup_data() {
    if [ -d "$DATA_DIR" ]; then
        echo "Backing up data directory to $BACKUP_DIR..."
        cp -r "$DATA_DIR"/* "$BACKUP_DIR/" 2>/dev/null || true
    fi
}

restore_data() {
    echo "Restoring data directory from backup..."
    rm -rf "$DATA_DIR"
    mkdir -p "$DATA_DIR"
    cp -r "$BACKUP_DIR"/* "$DATA_DIR/" 2>/dev/null || true
    rm -rf "$BACKUP_DIR"
}

cleanup() {
    restore_data
    rm -f /tmp/test_image.jpg
}

trap cleanup EXIT

echo -e "${BLUE}Testing Image Upload and Retrieval${NC}"
echo "====================================="
echo ""

# Backup existing data
backup_data

# Complete any active bake
if curl -s "$BASE_URL/status" | grep -q '"events":'; then
    echo "Completing active bake before test..."
    curl -s -X POST "$BASE_URL/log/loaf-complete" \
        -H "Content-Type: application/json" \
        -d '{"assessment":{"proof_level":"good","crumb_quality":5,"browning":"good","score":5,"notes":"Test completion"}}' > /dev/null 2>&1 || true
fi

# Create a small test JPEG image (1x1 red pixel)
echo -e "${BLUE}[TEST]${NC} Creating test image..."
base64 -d > /tmp/test_image.jpg << 'EOF'
/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0a
HBwgJC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/2wBDAQkJCQwLDBgNDRgyIRwhMjIyMjIy
MjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjL/wAARCAABAAEDASIA
AhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAv/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEB
AQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCwABmX
/9k=
EOF

if [ ! -f /tmp/test_image.jpg ]; then
    echo -e "${RED}✗ FAIL: Failed to create test image${NC}"
    exit 1
fi
echo -e "${GREEN}✓ PASS: Test image created${NC}"

# Start a new bake
echo ""
echo -e "${BLUE}[TEST]${NC} Starting new bake..."
response=$(curl -s -X POST "$BASE_URL/bake/start")
if echo "$response" | grep -q '"status":"loaf started"'; then
    echo -e "${GREEN}✓ PASS: Bake started${NC}"
else
    echo -e "${RED}✗ FAIL: Failed to start bake${NC}"
    exit 1
fi

# Upload note with image
echo ""
echo -e "${BLUE}[TEST]${NC} Uploading note with image..."
response=$(curl -s -X POST "$BASE_URL/log/note" \
    -F "note=Test image upload" \
    -F "image=@/tmp/test_image.jpg")

if echo "$response" | grep -q '"status":"logged"'; then
    echo -e "${GREEN}✓ PASS: Note with image uploaded${NC}"
else
    echo -e "${RED}✗ FAIL: Failed to upload note with image${NC}"
    echo "Response: $response"
    exit 1
fi

# Get current bake to find the image filename
echo ""
echo -e "${BLUE}[TEST]${NC} Retrieving bake data..."
bake_data=$(curl -s "$BASE_URL/api/bake/current")

if ! echo "$bake_data" | grep -q '"filename"'; then
    echo -e "${RED}✗ FAIL: Bake data missing filename field${NC}"
    exit 1
fi
echo -e "${GREEN}✓ PASS: Bake data has filename field${NC}"

# Extract filename and image name from the response
filename=$(echo "$bake_data" | grep -o '"filename":"[^"]*"' | cut -d'"' -f4)
image=$(echo "$bake_data" | grep -o '"image":"[^"]*"' | cut -d'"' -f4)

if [ -z "$filename" ] || [ -z "$image" ]; then
    echo -e "${RED}✗ FAIL: Could not extract filename or image from bake data${NC}"
    exit 1
fi

echo "Filename: $filename"
echo "Image: $image"

# Test image retrieval
echo ""
echo -e "${BLUE}[TEST]${NC} Retrieving uploaded image..."
image_url="/images/$filename/$image"
status_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL$image_url")

if [ "$status_code" = "200" ]; then
    echo -e "${GREEN}✓ PASS: Image retrieved successfully (HTTP 200)${NC}"
else
    echo -e "${RED}✗ FAIL: Image retrieval failed (HTTP $status_code)${NC}"
    exit 1
fi

# Verify image file exists on disk
echo ""
echo -e "${BLUE}[TEST]${NC} Verifying image file on disk..."
image_path="$DATA_DIR/images/$filename/$image"
if [ -f "$image_path" ]; then
    echo -e "${GREEN}✓ PASS: Image file exists at $image_path${NC}"
else
    echo -e "${RED}✗ FAIL: Image file not found at $image_path${NC}"
    exit 1
fi

# Test image-only note (no text)
echo ""
echo -e "${BLUE}[TEST]${NC} Uploading image-only note (no text)..."
response=$(curl -s -X POST "$BASE_URL/log/note" \
    -F "note=" \
    -F "image=@/tmp/test_image.jpg")

if echo "$response" | grep -q '"status":"logged"'; then
    echo -e "${GREEN}✓ PASS: Image-only note uploaded${NC}"
else
    echo -e "${RED}✗ FAIL: Failed to upload image-only note${NC}"
    echo "Response: $response"
    exit 1
fi

# Test that note without image or text fails
echo ""
echo -e "${BLUE}[TEST]${NC} Verifying empty note is rejected..."
status_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/log/note" \
    -F "note=")

if [ "$status_code" = "400" ]; then
    echo -e "${GREEN}✓ PASS: Empty note rejected (HTTP 400)${NC}"
else
    echo -e "${RED}✗ FAIL: Empty note should be rejected (got HTTP $status_code)${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}All image tests passed! ✓${NC}"
