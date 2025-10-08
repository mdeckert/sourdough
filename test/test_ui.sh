#!/bin/bash
# Test UI elements for the status view page
# This script tests that the page loads and contains expected elements

set -euo pipefail

BASE_URL="http://192.168.1.50:8080"
TEST_DATE="2025-10-05_08-00"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Testing Status View UI Elements${NC}"
echo "=================================="
echo ""

# Ensure test data exists
if [ ! -f "./data/bake_${TEST_DATE}.jsonl" ]; then
    echo "Generating test data..."
    ./test/generate_test_data.sh > /dev/null
fi

echo "Fetching status page HTML..."
HTML=$(curl -s "${BASE_URL}/view/status?date=${TEST_DATE}")

# Test 1: Page loads
if [ -z "$HTML" ]; then
    echo -e "${RED}✗ FAIL: Page did not load${NC}"
    exit 1
fi
echo -e "${GREEN}✓ PASS: Page loaded successfully${NC}"

# Test 2: Chart canvas exists
if echo "$HTML" | grep -q 'id="tempChart"'; then
    echo -e "${GREEN}✓ PASS: Temperature chart canvas found${NC}"
else
    echo -e "${RED}✗ FAIL: Temperature chart canvas not found${NC}"
    exit 1
fi

# Test 3: Reset Zoom button exists
if echo "$HTML" | grep -q 'onclick="resetZoom()"'; then
    echo -e "${GREEN}✓ PASS: Reset Zoom button found${NC}"
else
    echo -e "${RED}✗ FAIL: Reset Zoom button not found${NC}"
    exit 1
fi

# Test 4: Zoom to Baking button exists
if echo "$HTML" | grep -q 'onclick="zoomToBaking()"'; then
    echo -e "${GREEN}✓ PASS: Zoom to Baking button found${NC}"
else
    echo -e "${RED}✗ FAIL: Zoom to Baking button not found${NC}"
    exit 1
fi

# Test 5: Delete button exists
if echo "$HTML" | grep -q 'onclick="deleteBake()"'; then
    echo -e "${GREEN}✓ PASS: Delete button found${NC}"
else
    echo -e "${RED}✗ FAIL: Delete button not found${NC}"
    exit 1
fi

# Test 6: Chart.js zoom plugin included
if echo "$HTML" | grep -q 'chartjs-plugin-zoom'; then
    echo -e "${GREEN}✓ PASS: Chart.js zoom plugin included${NC}"
else
    echo -e "${RED}✗ FAIL: Chart.js zoom plugin not found${NC}"
    exit 1
fi

# Test 7: resetZoom function exists
if echo "$HTML" | grep -q 'function resetZoom()'; then
    echo -e "${GREEN}✓ PASS: resetZoom function found${NC}"
else
    echo -e "${RED}✗ FAIL: resetZoom function not found${NC}"
    exit 1
fi

# Test 8: zoomToBaking function exists
if echo "$HTML" | grep -q 'function zoomToBaking()'; then
    echo -e "${GREEN}✓ PASS: zoomToBaking function found${NC}"
else
    echo -e "${RED}✗ FAIL: zoomToBaking function not found${NC}"
    exit 1
fi

# Test 9: Check for datasets (Kitchen, Dough, Loaf, Oven, Notes)
if echo "$HTML" | grep -q "label: 'Kitchen Temp"; then
    echo -e "${GREEN}✓ PASS: Kitchen temperature dataset found${NC}"
else
    echo -e "${RED}✗ FAIL: Kitchen temperature dataset not found${NC}"
    exit 1
fi

if echo "$HTML" | grep -q "label: 'Dough Temp"; then
    echo -e "${GREEN}✓ PASS: Dough temperature dataset found${NC}"
else
    echo -e "${RED}✗ FAIL: Dough temperature dataset not found${NC}"
    exit 1
fi

if echo "$HTML" | grep -q "label: 'Loaf Internal Temp"; then
    echo -e "${GREEN}✓ PASS: Loaf internal temperature dataset found${NC}"
else
    echo -e "${RED}✗ FAIL: Loaf internal temperature dataset not found${NC}"
    exit 1
fi

if echo "$HTML" | grep -q "label: 'Oven Temp"; then
    echo -e "${GREEN}✓ PASS: Oven temperature dataset found${NC}"
else
    echo -e "${RED}✗ FAIL: Oven temperature dataset not found${NC}"
    exit 1
fi

if echo "$HTML" | grep -q "label: 'Notes'"; then
    echo -e "${GREEN}✓ PASS: Notes dataset found${NC}"
else
    echo -e "${RED}✗ FAIL: Notes dataset not found${NC}"
    exit 1
fi

# Test 10: Timeline exists
if echo "$HTML" | grep -q 'id="timeline"'; then
    echo -e "${GREEN}✓ PASS: Timeline element found${NC}"
else
    echo -e "${RED}✗ FAIL: Timeline element not found${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}All UI element tests passed! ✓${NC}"
echo ""
echo "To manually test:"
echo "  1. Open: ${BASE_URL}/view/status?date=${TEST_DATE}"
echo "  2. Verify chart displays 5 datasets:"
echo "     - Kitchen Temp (blue, before oven-in)"
echo "     - Dough Temp (red, before oven-in)"
echo "     - Loaf Internal Temp (purple, during baking)"
echo "     - Oven Temp (orange, during baking)"
echo "     - Notes (yellow markers)"
echo "  3. Click 'Zoom to Baking Phase' - should show ~10:30-11:10, Y-axis ~160-520°F"
echo "  4. Click 'Reset Zoom' - should restore full view (60-500°F)"
echo "  5. Hover over yellow note markers - should show note text in tooltip"
echo "  6. No long line connecting dough (70-76°F) to loaf temps (180-208°F)"
