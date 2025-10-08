# Claude Code Context - Sourdough Logger

## Project Overview
This is a sourdough bread logging system that tracks the entire baking process from starter feeding to final assessment. It uses QR codes for quick event logging via mobile device and provides a web UI for detailed data entry.

**IMPORTANT - Server Access**:
- Server IP: `192.168.1.50:8080`
- **Always use the server IP address** (not localhost) in documentation, examples, and commands
- QR codes must use the server IP to work on mobile devices
- Localhost/127.0.0.1 are only for local testing on the server itself

## Architecture

### Core Components
```
sourdough/
├── cmd/
│   ├── server/     - HTTP server (port 8080)
│   ├── sourdough/  - CLI tool for viewing/analyzing data
│   └── qrgen/      - QR code PDF generator
├── internal/
│   ├── server/     - HTTP handlers and web UI templates
│   ├── storage/    - JSONL file-based storage
│   ├── models/     - Event and bake data structures
│   └── qr/         - QR code generation logic
├── data/           - JSONL bake files (bake_YYYY-MM-DD_HH-MM.jsonl)
├── qrcodes/        - Generated QR codes PDF
└── test/           - Integration tests

```

### Storage Format
- **JSONL files**: One line per event, stored in `data/`
- **File naming**: `bake_2025-10-07_19-06.jsonl` (includes timestamp)
- **Multi-day support**: Bakes continue across days until marked complete
- **Active bake**: Most recent file without `bake-complete` event

### Event Types
1. `starter-out` - Start new bake
2. `fed` - Feed starter
3. `levain-ready` - Levain doubled
4. `mixed` - Mix dough
5. `fold` - Fold dough (auto-counts: 1, 2, 3...)
6. `shaped` - Shape loaf
7. `fridge-in` - Begin cold proof
8. `oven-in` - Start baking (implies fridge-out)
9. `oven-out` - End baking, loaf cooling
10. `loaf-complete` - Finish with assessment
11. `temperature` - Log temp (kitchen, dough, or oven)
12. `note` - Add text observation

**Note**: There is no `fridge-out` event - use `oven-in` as it marks the transition from cold proof to baking.

## Build & Deploy

### Build Commands
```bash
# Build all binaries
make build          # Builds server, CLI, and qrgen

# Build individually
make server         # → bin/sourdough-server
make cli            # → bin/sourdough
make qrgen          # → bin/qrgen
```

### Deploy Server
```bash
# Build and update running service
make server
sudo systemctl restart sourdough

# Check status
sudo systemctl status sourdough

# View logs
journalctl -u sourdough -f
```

**IMPORTANT**: Always use `make server` to build, NOT `go build` directly. The Makefile ensures the binary goes to `bin/sourdough-server` which is what the systemd service expects.

### Service Configuration
- **Service file**: `/etc/systemd/system/sourdough.service`
- **Binary path**: `/home/mdeckert/sourdough/bin/sourdough-server`
- **Data directory**: `/home/mdeckert/sourdough/data`
- **Port**: 8080
- **Auto-restart**: Yes (on failure)

## Testing

### Test Suite Overview
- **26 unit tests** (handlers + storage + QR generator)
- **33 integration tests** (full workflow with multiple instances)
- **Coverage**: ~61-66%
- **Data safety**: Integration tests backup and restore data automatically
- **QR validation**: Tests ensure QR codes don't point to localhost

### Running Tests
```bash
# Unit tests (fast, ~30ms)
make test

# Coverage report (generates coverage.html)
make test-coverage

# Generate test dataset (comprehensive 2-day bake)
make test-data

# Integration tests (requires server running, backs up & restores data)
make test-integration

# All tests
make test-all
```

**Test Data Behavior**:
- Integration tests check for test dataset (`bake_2025-10-05_08-00.jsonl`)
- If missing: auto-generates and removes after tests complete
- If present: uses existing data and preserves it
- Use `make test-data` to manually generate/regenerate test dataset

### Before Making Changes
1. Run `make test` to establish baseline
2. Make your changes
3. Run `make test` again to catch regressions
4. Run `make server && sudo systemctl restart sourdough`
5. Run `make test-integration` to verify end-to-end

## API Endpoints

### Core Endpoints
- `POST /bake/start` - Start new bake (prevents duplicates)
- `POST /log/{event}` - Log event (e.g., /log/fold, /log/mixed)
- `POST /log/temp/{value}` - Log temperature
  - Query param `?type=dough` for dough temp (default: kitchen)
- `POST /log/note` - Log note (JSON body: `{"note":"text"}`)
- `POST /log/bake-complete` - Complete with assessment (JSON body)
- `GET /status` - Get current bake status

### Web UI Pages
- `GET /temp` - Temperature entry form (slider 60-80°F)
- `GET /notes` - Note entry form
- `GET /complete` - Bake assessment form
- `GET /qrcodes.pdf` - Download QR codes PDF

### QR Code Behavior
- All QR endpoints accept both GET and POST
- GET requests show browser-friendly HTML responses
- POST requests return JSON (for API usage)

## Key Design Decisions

### 1. Multi-Day Bakes
**Problem**: Original design created one file per day, breaking overnight bakes.

**Solution**:
- Files named with timestamp: `bake_YYYY-MM-DD_HH-MM.jsonl`
- System finds most recent file without `bake-complete` event
- Events continue in same file across multiple days
- Only `bake-complete` event allows starting new bake

### 2. Fold Counting
Folds auto-increment by checking the last event:
- If last event is `fold` with `fold_count: N`, next fold is `N+1`
- Otherwise, fold count is `1`
- No manual counting needed

### 3. Temperature Storage
Two separate fields to track different temps:
- `temp_f`: Kitchen/ambient temperature
- `dough_temp_f`: Dough temperature
- Use `?type=dough` query param to specify dough temp

### 4. QR Code URL Validation
**Problem**: QR codes pointing to localhost don't work on mobile devices.

**Solution**:
- `qrgen` validates URL and rejects localhost/127.0.0.1
- Tests ensure URLs use proper IP addresses or hostnames
- Error message guides users to use server IP (e.g., 192.168.1.50:8080)

### 5. Assessments
Stored as part of `loaf-complete` event's `data` field:
```json
{
  "event": "loaf-complete",
  "data": {
    "assessment": {
      "proof_level": "good|underproofed|overproofed",
      "crumb_quality": 1-10,
      "browning": "none|slight|good|over",
      "score": 1-10,
      "notes": "optional text"
    }
  }
}
```

## Common Tasks

### Generate QR Codes
```bash
./bin/qrgen http://YOUR_SERVER_IP:8080
# Example: ./bin/qrgen http://192.168.1.50:8080
# Creates qrcodes/qrcodes.pdf with all 15 QR codes
#
# IMPORTANT: Must use server IP, not localhost!
# Tool will reject localhost URLs to prevent broken QR codes
```

### View Bake History
```bash
./bin/sourdough history
./bin/sourdough history --limit 5
```

### Check Current Bake
```bash
./bin/sourdough status
curl http://192.168.1.50:8080/status
```

### Generate Test Dataset
```bash
./test/generate_test_data.sh
# Creates comprehensive 2-day test bake: data/bake_2025-10-05_08-00.jsonl
# View at: http://192.168.1.50:8080/view/status?date=2025-10-05_08-00
```

### Delete a Bake
Bakes can be deleted via the UI (delete button at bottom of status page) or via API:
```bash
# Delete via API (moves to data/trash/)
curl -X DELETE http://192.168.1.50:8080/api/bake/2025-10-05_08-00

# Deleted files are moved to trash, not permanently removed
ls data/trash/
```

### Manual Event Logging (for testing)
```bash
# Start bake
curl -X POST http://192.168.1.50:8080/loaf/start

# Log temperature
curl -X POST http://192.168.1.50:8080/log/temp/72
curl -X POST "http://192.168.1.50:8080/log/temp/76?type=dough"

# Log event
curl -X POST http://192.168.1.50:8080/log/fold

# Log note
curl -X POST http://192.168.1.50:8080/log/note \
  -H "Content-Type: application/json" \
  -d '{"note":"Good oven spring"}'
```

## Recent Changes

### Temperature UI Update (2025-10-07)
Changed from quick-select buttons to slider interface:
- **Before**: 6 preset buttons + manual entry + separate submit button
- **After**: Slider (60-80°F) + manual entry + direct submit via Dough/Kitchen buttons
- **Files changed**: `internal/server/templates.go`
- **Why**: Faster workflow, cleaner UI, better mobile experience

### Test Suite Addition (2025-10-07)
Added comprehensive automated testing:
- **Unit tests**: `internal/server/handlers_test.go`, `internal/storage/jsonl_test.go`
- **Integration**: `test/integration_test.sh`
- **Makefile targets**: `test`, `test-coverage`, `test-integration`, `test-all`
- **Coverage**: 61-66% code coverage

## Troubleshooting

### Server Won't Start
```bash
# Check if already running
sudo systemctl status sourdough

# Check logs
journalctl -u sourdough -n 50

# Verify binary exists
ls -la /home/mdeckert/sourdough/bin/sourdough-server

# Rebuild and restart
make server
sudo systemctl restart sourdough
```

### Tests Failing
```bash
# Check if server is running (for integration tests)
curl http://192.168.1.50:8080/health

# Clean and rebuild
make clean
make build
make test

# Check for active bake (can cause duplicate start errors)
curl http://192.168.1.50:8080/status
```

### QR Codes Not Working
```bash
# Regenerate QR codes
./bin/qrgen http://192.168.1.50:8080

# Verify PDF exists
ls -la qrcodes/qrcodes.pdf

# Test PDF endpoint
curl -I http://192.168.1.50:8080/qrcodes.pdf
```

## Development Workflow

### Adding a New Event Type
1. Add to `internal/models/events.go` constants
2. Add validation in `internal/server/handlers.go` validEvents map
3. Add test case in `internal/server/handlers_test.go`
4. Add to integration test in `test/integration_test.sh`
5. Update QR code generation if needed
6. Run `make test && make server && sudo systemctl restart sourdough`

### Modifying Web UI
1. Edit template in `internal/server/templates.go` (inline HTML/CSS/JS)
2. Run `make server` to rebuild
3. Run `sudo systemctl restart sourdough` to deploy
4. Test in browser at http://192.168.1.50:8080/{page}
5. Verify with integration tests: `make test-integration`

### Storage Changes
1. Modify `internal/storage/jsonl.go`
2. Update/add tests in `internal/storage/jsonl_test.go`
3. Ensure backward compatibility with existing JSONL files
4. Run `make test` before deploying
5. Consider data migration if format changes

## Important Notes

### DO NOT
- ❌ Use `go build ./cmd/server` directly (wrong output path)
- ❌ Edit files in `/etc/systemd/system/` without `systemctl daemon-reload`
- ❌ Delete `data/` directory (contains all bake history)
- ❌ Change JSONL format without migration plan
- ❌ Skip tests before deploying changes

### ALWAYS
- ✅ Use `make server` to build
- ✅ Run `make test` after changes
- ✅ Use `sudo systemctl restart sourdough` to deploy
- ✅ Check `sudo systemctl status sourdough` after deploy
- ✅ Test with `make test-integration` for major changes
- ✅ Keep JSONL files (they're your data!)

## File Locations Reference

### Binaries
- Server: `bin/sourdough-server`
- CLI: `bin/sourdough`
- QR Generator: `bin/qrgen`

### Data
- Bake files: `data/bake_*.jsonl`
- Deleted bakes: `data/trash/bake_*.jsonl` (moved here when deleted via UI/API)
- QR codes: `qrcodes/qrcodes.pdf`
- Test data generator: `test/generate_test_data.sh`

### Config
- Systemd service: `/etc/systemd/system/sourdough.service`
- No other config files (uses environment variables)

### Tests
- Unit: `internal/*/\*_test.go`
- Integration: `test/integration_test.sh`
- Coverage: `coverage.out`, `coverage.html` (generated)

### Documentation
- This file: `CLAUDE.md` (context for Claude Code)
- Test docs: `TEST_SUITE.txt`
- Updates: `MAJOR_UPDATE.txt`, `TEMP_UI_UPDATE.txt`, etc.

## Quick Reference Card

```bash
# Build & Deploy
make server && sudo systemctl restart sourdough

# Test Everything
make test-all

# Check Status
sudo systemctl status sourdough
curl http://192.168.1.50:8080/health

# View Logs
journalctl -u sourdough -f

# Generate QR Codes
./bin/qrgen http://192.168.1.50:8080
```

---

**Last Updated**: 2025-10-07
**Project Version**: 1.0 (with test suite)
**Go Version**: 1.21+
**Platform**: Linux (systemd)
