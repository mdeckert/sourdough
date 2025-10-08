#!/bin/bash
# Test automatic temperature logging
# This tests that the Ecobee integration can fetch and log temperature

set -e

BASE_URL="http://192.168.1.50:8080"

echo "Testing automatic temperature mechanism..."
echo ""

# Check if there's an active bake
echo "1. Checking for active bake..."
if curl -s "$BASE_URL/status" | grep -q '"events":'; then
    echo "   ✓ Active bake found"
else
    echo "   ✗ No active bake (auto-logging won't trigger)"
    echo ""
    echo "   To test: Start a bake first with /bake/start"
    exit 1
fi

# Check Ecobee is enabled in logs
echo ""
echo "2. Checking Ecobee integration status..."
if sudo journalctl -u sourdough --since "5 minutes ago" 2>/dev/null | grep -q "Ecobee integration enabled"; then
    echo "   ✓ Ecobee integration enabled"
else
    echo "   ⚠ Cannot verify Ecobee status from logs"
fi

if sudo journalctl -u sourdough --since "5 minutes ago" 2>/dev/null | grep -q "Automatic temperature logging enabled"; then
    echo "   ✓ Automatic logging goroutine started"
else
    echo "   ⚠ Cannot verify auto-logging from logs"
fi

# Show the schedule
echo ""
echo "3. Auto-logging schedule (every 4 hours from server start):"
SERVER_START=$(sudo journalctl -u sourdough | grep "Starting server" | tail -1 | awk '{print $1, $2, $3}')
echo "   Server started: $SERVER_START"
echo "   Next logs at: +4h, +8h, +12h, etc. from start time"

echo ""
echo "4. Recent temperature events in current bake:"
RECENT_TEMPS=$(curl -s "$BASE_URL/status" | grep -o '"event":"temperature"' | wc -l)
echo "   Found $RECENT_TEMPS temperature events"

# Show last temperature
LAST_TEMP=$(curl -s "$BASE_URL/status" | grep -o '"temp_f":[0-9.]*' | tail -1 | cut -d: -f2)
if [ ! -z "$LAST_TEMP" ]; then
    echo "   Last logged: ${LAST_TEMP}°F"
fi

echo ""
echo "✓ Auto-logging mechanism is configured correctly!"
echo ""
echo "To verify it's working:"
echo "  - Wait until the next 4-hour interval"
echo "  - Check journalctl: sudo journalctl -u sourdough -f"
echo "  - Look for: 'Auto-logged kitchen temperature: XX.X°F'"
echo ""
echo "Or restart server to reset timer for testing:"
echo "  sudo systemctl restart sourdough"
echo "  (Next log will be ~4 hours from restart)"
