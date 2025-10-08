#!/bin/bash
# Integration test suite for Sourdough Logger
# Tests the full workflow from starting a bake to completion

set -euo pipefail

BASE_URL="http://192.168.1.50:8080"
DATA_DIR="./data"
BACKUP_DIR=$(mktemp -d)
TEMP_DIR=$(mktemp -d)
TEST_DATA_FILE="$DATA_DIR/bake_2025-10-05_08-00.jsonl"
TEST_DATA_GENERATED=false

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
    rm -rf "$TEMP_DIR"

    # If we generated test data, remove it
    if [ "$TEST_DATA_GENERATED" = true ] && [ -f "$TEST_DATA_FILE" ]; then
        echo "Removing auto-generated test data..."
        rm -f "$TEST_DATA_FILE"
    fi
}

# Set up traps
trap "restore_data; cleanup" EXIT

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
PASSED=0
FAILED=0

# Helper functions
log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    PASSED=$((PASSED + 1))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    FAILED=$((FAILED + 1))
}

test_health_check() {
    log_test "Health check"

    response=$(curl -s "$BASE_URL/health")
    if echo "$response" | grep -q '"status":"ok"'; then
        log_pass "Health check returned OK"
    else
        log_fail "Health check failed"
    fi
}

test_start_bake() {
    log_test "Start new bake"

    response=$(curl -s -X POST "$BASE_URL/bake/start")

    if echo "$response" | grep '"status":"loaf started"' > /dev/null 2>&1; then
        log_pass "Bake started successfully"
    elif echo "$response" | grep "already started" > /dev/null 2>&1; then
        log_pass "Bake already active (continuing with existing bake)"
    else
        log_fail "Failed to start bake: $response"
    fi
}

test_duplicate_start() {
    log_test "Prevent duplicate bake start"

    status_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/bake/start")
    if [ "$status_code" = "400" ]; then
        log_pass "Correctly prevented duplicate bake start"
    else
        log_fail "Should have prevented duplicate start (got status $status_code)"
    fi
}

test_log_events() {
    log_test "Log workflow events"

    events=("fed" "levain-ready" "mixed" "fold" "fold" "fold" "fold" "shaped" "fridge-in" "fridge-out" "oven-in")

    for event in "${events[@]}"; do
        response=$(curl -s -X POST "$BASE_URL/log/$event")
        if echo "$response" | grep -q '"status":"logged"'; then
            log_pass "Event '$event' logged"
        else
            log_fail "Failed to log event '$event': $response"
        fi
        sleep 0.05
    done
}

test_temperature_logging() {
    log_test "Log multiple temperatures"

    # Multiple kitchen temps
    kitchen_temps=(68 70 72 74)
    for temp in "${kitchen_temps[@]}"; do
        response=$(curl -s -X POST "$BASE_URL/log/temp/$temp")
        if echo "$response" | grep -q "\"temp_f\":$temp"; then
            log_pass "Kitchen temperature ${temp}°F logged"
        else
            log_fail "Failed to log kitchen temp $temp: $response"
        fi
        sleep 0.05
    done

    # Multiple dough temps
    dough_temps=(75 76 77)
    for temp in "${dough_temps[@]}"; do
        response=$(curl -s -X POST "$BASE_URL/log/temp/$temp?type=dough")
        if echo "$response" | grep -q "\"dough_temp_f\":$temp"; then
            log_pass "Dough temperature ${temp}°F logged"
        else
            log_fail "Failed to log dough temp $temp: $response"
        fi
        sleep 0.05
    done
}

test_note_logging() {
    log_test "Log note"

    response=$(curl -s -X POST "$BASE_URL/log/note" \
        -H "Content-Type: application/json" \
        -d '{"note":"Good oven spring, nice crust"}')

    if echo "$response" | grep -q '"status":"logged"'; then
        log_pass "Note logged successfully"
    else
        log_fail "Failed to log note: $response"
    fi
}

test_fold_counting() {
    log_test "Fold counting and verification"

    # Get status to check fold count
    response=$(curl -s "$BASE_URL/status")

    # Count fold events (should be 4 from earlier test)
    fold_count=$(echo "$response" | grep -o '"event":"fold"' | wc -l)
    if [ "$fold_count" -eq 4 ]; then
        log_pass "Fold count correct (4 folds)"
    else
        log_fail "Expected 4 folds, got $fold_count"
    fi

    # Verify fold_count values increment (1, 2, 3, 4)
    if echo "$response" | grep -q '"fold_count":1' && \
       echo "$response" | grep -q '"fold_count":2' && \
       echo "$response" | grep -q '"fold_count":3' && \
       echo "$response" | grep -q '"fold_count":4'; then
        log_pass "Fold counts increment correctly (1,2,3,4)"
    else
        log_fail "Fold counts not incrementing properly"
    fi
}

test_bake_completion() {
    log_test "Complete bake with assessment"

    response=$(curl -s -X POST "$BASE_URL/log/loaf-complete" \
        -H "Content-Type: application/json" \
        -d '{
            "assessment": {
                "proof_level": "good",
                "crumb_quality": 8,
                "browning": "good",
                "score": 9,
                "notes": "Excellent loaf"
            }
        }')

    if echo "$response" | grep -q '"status":"logged"'; then
        log_pass "Bake completed with assessment"
    else
        log_fail "Failed to complete bake: $response"
    fi
}

test_new_bake_after_completion() {
    log_test "Start new bake after completion"

    # Small delay to ensure timestamp difference
    sleep 0.2

    response=$(curl -s -X POST "$BASE_URL/bake/start")
    if echo "$response" | grep -q '"status":"loaf started"'; then
        log_pass "New bake started after completion"
    else
        log_fail "Failed to start new bake after completion: $response"
    fi
}

test_web_ui_pages() {
    log_test "Web UI pages accessibility"

    pages=("/temp" "/notes" "/complete")

    for page in "${pages[@]}"; do
        status_code=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL$page")
        if [ "$status_code" = "200" ]; then
            log_pass "Page $page accessible"
        else
            log_fail "Page $page returned status $status_code"
        fi
    done
}

test_status_endpoint() {
    log_test "Status endpoint"

    response=$(curl -s "$BASE_URL/status")
    if echo "$response" | grep -q '"events"'; then
        log_pass "Status endpoint returns events"
    else
        log_fail "Status endpoint missing events: $response"
    fi
}

test_invalid_requests() {
    log_test "Invalid request handling"

    # Invalid temperature
    status_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/log/temp/invalid")
    if [ "$status_code" = "400" ]; then
        log_pass "Invalid temperature rejected"
    else
        log_fail "Should reject invalid temperature (got $status_code)"
    fi

    # Empty note
    status_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/log/note" \
        -H "Content-Type: application/json" \
        -d '{"note":""}')
    if [ "$status_code" = "400" ]; then
        log_pass "Empty note rejected"
    else
        log_fail "Should reject empty note (got $status_code)"
    fi

    # Invalid event
    status_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/log/invalid-event")
    if [ "$status_code" = "400" ]; then
        log_pass "Invalid event rejected"
    else
        log_fail "Should reject invalid event (got $status_code)"
    fi
}

# Main test execution
echo "========================================="
echo "Sourdough Logger Integration Tests"
echo "========================================="
echo ""
echo "This test suite will:"
echo "  • Backup current data directory"
echo "  • Test all endpoints and workflows"
echo "  • Restore original data when complete"
echo ""
echo "Tests covered:"
echo "  ✓ Health check endpoint"
echo "  ✓ Bake start (new & duplicate prevention)"
echo "  ✓ 11 workflow events (fed, levain-ready, mixed, 4 folds, shaped, fridge-in/out, oven-in)"
echo "  ✓ 7 temperature logs (4 kitchen, 3 dough)"
echo "  ✓ Note logging"
echo "  ✓ Fold count auto-increment (1→2→3→4)"
echo "  ✓ Bake completion with assessment"
echo "  ✓ New bake after completion"
echo "  ✓ Web UI pages (/temp, /notes, /complete)"
echo "  ✓ Status endpoint"
echo "  ✓ Invalid requests (bad temp, empty note, invalid event)"
echo ""
echo "========================================="
echo ""

# Check if server is running
if ! curl -s "$BASE_URL/health" > /dev/null 2>&1; then
    echo -e "${RED}ERROR:${NC} Server not running at $BASE_URL"
    echo "Start the server with: make server && ./bin/sourdough-server"
    exit 1
fi

# Backup data
backup_data
echo ""

# Run tests in order
test_health_check
test_start_bake
test_duplicate_start
test_log_events
test_temperature_logging
test_note_logging
test_fold_counting
test_bake_completion
test_new_bake_after_completion
test_web_ui_pages
test_status_endpoint
test_invalid_requests

# Test 9: View pages with test dataset
test_view_pages() {
    log_test "Testing view pages with comprehensive dataset"

    # Check if test data exists, generate if missing
    if [ ! -f "$TEST_DATA_FILE" ]; then
        log_test "Test dataset not found, generating..."
        if [ -f "./test/generate_test_data.sh" ]; then
            ./test/generate_test_data.sh > /dev/null
            TEST_DATA_GENERATED=true
            log_test "Test data generated (will be removed after tests)"
        else
            log_fail "Test data generator script not found"
            return
        fi
    else
        log_test "Using existing test dataset (will be preserved)"
    fi

    # Test API endpoint for specific bake
    log_test "Testing API endpoint for specific bake"
    RESPONSE=$(curl -s "$BASE_URL/api/bake/2025-10-05_08-00")

    # Check that response contains expected data
    if ! echo "$RESPONSE" | grep -q '"event":"oven-in"'; then
        log_fail "API response missing oven-in event"
        return
    fi

    if ! echo "$RESPONSE" | grep -q '"event":"oven-out"'; then
        log_fail "API response missing oven-out event"
        return
    fi

    # Check for temperature data during baking
    if ! echo "$RESPONSE" | grep -q '"dough_temp_f":208'; then
        log_fail "API response missing loaf temp readings"
        return
    fi

    # Test status view page returns HTML
    log_test "Testing status view page"
    STATUS_HTML=$(curl -s "$BASE_URL/view/status?date=2025-10-05_08-00")

    if ! echo "$STATUS_HTML" | grep -q "Bake Status"; then
        log_fail "Status view page doesn't contain expected title"
        return
    fi

    if ! echo "$STATUS_HTML" | grep -q "zoomToBaking"; then
        log_fail "Status view page missing zoom to baking function"
        return
    fi

    # Test history view page
    log_test "Testing history view page"
    HISTORY_HTML=$(curl -s "$BASE_URL/view/history")

    if ! echo "$HISTORY_HTML" | grep -q "Bake History"; then
        log_fail "History view page doesn't contain expected title"
        return
    fi

    # Test that bakes list API includes our test bake
    log_test "Testing bakes list API"
    BAKES_LIST=$(curl -s "$BASE_URL/api/bakes")

    if ! echo "$BAKES_LIST" | grep -q "2025-10-05_08-00"; then
        log_fail "Bakes list doesn't include test bake"
        return
    fi

    log_pass "View pages and APIs working correctly"
}

test_view_pages

# Summary
echo ""
echo "========================================="
echo "Test Results"
echo "========================================="
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo "========================================="

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed! ✓${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed! ✗${NC}"
    exit 1
fi
