# Testing Guide

## Test Data Management

**All tests automatically preserve your current loaf data.**

Tests that create real data (integration and QR tests) use backup/restore to prevent pollution:

```bash
# Your data before test
data/bake_2025-10-07_19-13-49.jsonl

# Test runs (creates temporary data)
# ...test creates events...

# Your data after test (unchanged!)
data/bake_2025-10-07_19-13-49.jsonl
```

## Test Types

### 1. Unit Tests (`make test`)
- **Tests**: Go unit tests for handlers and storage
- **Data**: Uses temporary directories, no impact on your data
- **Speed**: Fast (~30ms)
- **Safe**: ✅ Never touches real data

### 2. Integration Tests (`make test-integration`)
- **Tests**: Full workflow with real server
- **Data**: ✅ **Backs up and restores data automatically**
- **Speed**: Moderate (~5 seconds)
- **Safe**: ✅ Your loaf data is preserved

### 3. UI Tests (`make test-ui`)
- **Tests**: Status page rendering and chart elements
- **Data**: Uses existing test data file (`bake_2025-10-05_08-00.jsonl`)
- **Speed**: Fast (~1 second)
- **Safe**: ✅ Only reads data, never writes

### 4. QR Code Tests (`make test-qr`)
- **Tests**: QR generation, URL validation, endpoint verification
- **Data**: ✅ **Backs up and restores data automatically**
- **Speed**: Fast (~2 seconds)
- **Safe**: ✅ Your loaf data is preserved

### 5. All Tests (`make test-all`)
- **Runs**: All 4 test suites in order
- **Data**: ✅ **All data is preserved**
- **Time**: ~10 seconds total

## How Backup/Restore Works

Tests that create real data use this pattern:

```bash
# 1. Backup
cp -r data/* /tmp/backup/

# 2. Run test (may create new data)
curl -X POST http://localhost:8080/log/mixed

# 3. Restore (even if test fails)
rm -rf data/*
cp -r /tmp/backup/* data/
```

This is implemented with bash `trap` to ensure restoration even on test failure.

## Running Tests Safely

```bash
# Run any test - your data is safe
make test
make test-integration
make test-ui
make test-qr
make test-all

# Your current loaf is never affected
```

## Test Data Files

- **Test data**: `data/bake_2025-10-05_08-00.jsonl` (used by UI tests)
- **Your data**: Any other files in `data/` (preserved by all tests)
- **Backups**: Tests use `/tmp/sourdough-*-backup-*` (cleaned up automatically)

## What Each Test Does

### Unit Tests
```bash
make test
# Tests:
# ✓ Health check
# ✓ Bake start/duplicate prevention
# ✓ Temperature logging
# ✓ Note logging
# ✓ All workflow events
# ✓ Fold counting
# ✓ Bake completion
# ✓ Status endpoint
# ✓ QR PDF serving
# ✓ Web UI pages
# ✓ HTTP method validation
```

### Integration Tests
```bash
make test-integration
# Tests full workflow:
# ✓ Start bake
# ✓ Fed → Levain Ready → Mixed
# ✓ 4 Folds (with auto-counting)
# ✓ Shaped → Fridge In → Fridge Out
# ✓ Oven In → Bake Complete
# ✓ Temperature and note logging
# ✓ Status queries
# ✓ Duplicate bake prevention
```

### UI Tests
```bash
make test-ui
# Tests status page:
# ✓ Page loads
# ✓ Contains timeline
# ✓ Contains chart
# ✓ Has 5 temperature datasets
# ✓ Zoom and reset functions exist
```

### QR Code Tests
```bash
make test-qr
# Tests QR generation:
# ✓ Rejects --help flag
# ✓ Rejects localhost URLs
# ✓ Generates all QR code files
# ✓ QR code endpoints work
# ✓ Restores original data
```

## Test Coverage

Current coverage: **~61-66%**

```bash
# Generate coverage report
make test-coverage
# Opens: coverage.html
```

## Adding New Tests

When adding tests that create data:

1. **Add backup/restore**:
   ```bash
   BACKUP_DIR="/tmp/mytest-backup-$$"
   cp -r data/* "$BACKUP_DIR/"
   trap "cp -r $BACKUP_DIR/* data/; rm -rf $BACKUP_DIR" EXIT
   ```

2. **Run your test** (creates data)

3. **Cleanup happens automatically** (via trap)

## Troubleshooting

### "My data got polluted by tests"

**This shouldn't happen anymore!** All tests now preserve data.

If it does happen:
1. Check which test you ran
2. Report the issue (it's a bug)
3. Your data might be in `/tmp/sourdough-*-backup-*`

### "Tests are failing"

**Unit tests:**
```bash
# Check server code
make test
```

**Integration tests:**
```bash
# Ensure server is running
sudo systemctl status sourdough

# Check health
curl http://localhost:8080/health
```

**UI tests:**
```bash
# Ensure test data exists
ls data/bake_2025-10-05_08-00.jsonl

# Regenerate if missing
./test/generate_test_data.sh
```

**QR tests:**
```bash
# Ensure qrgen is built
make qrgen

# Ensure server is running
sudo systemctl status sourdough
```

## CI/CD Integration

All tests are safe for CI/CD:

```bash
# Run everything
make test-all

# Exit code 0 = pass
# Exit code 1 = fail
```

No manual data cleanup needed before or after tests.
