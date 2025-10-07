# Sourdough Bread Logging System

A lightweight Go-based system for logging and analyzing sourdough baking process.

## Installation

### Install Go

```bash
# Download and install Go 1.21+
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### Build Binaries

```bash
cd /home/mdeckert/sourdough

# Build server
go build -o bin/sourdough-server ./cmd/server

# Build CLI tool
go build -o bin/sourdough ./cmd/sourdough

# Add CLI to PATH (optional)
sudo ln -s $(pwd)/bin/sourdough /usr/local/bin/sourdough
```

## Usage

### Start the Server

```bash
# Run in foreground (for testing)
./bin/sourdough-server

# Or install as systemd service (see sourdough.service)
sudo cp sourdough.service /etc/systemd/system/
sudo systemctl enable sourdough
sudo systemctl start sourdough
```

### CLI Commands

```bash
# Start a new bake
sourdough start

# Log events
sourdough log starter-out
sourdough log fed
sourdough log mixed
sourdough log fold
sourdough log shaped
sourdough log fridge-in
sourdough log oven-in

# Log temperatures
sourdough temp 76
sourdough log temp 76 --dough  # dough temp specifically

# Check current status
sourdough status

# Complete bake with assessment
sourdough complete

# View history
sourdough history
sourdough review 2025-10-07
```

### QR Code Logging

Generate QR codes for quick phone-based logging:

```bash
# Generate all QR codes
sourdough qr generate

# QR codes will be in ./qrcodes/ directory
# Print qrcodes/sheet.png and stick on fridge
```

## Data Storage

- Bakes stored in `./data/` as JSON Lines files
- One file per bake: `bake_YYYY-MM-DD.jsonl`
- Each line is a timestamped event in JSON format
- Human-readable and easy to backup/analyze

## Architecture

- **Server**: Lightweight HTTP server (port 8080) for receiving log events
- **CLI**: Command-line tool for interactive logging and analysis
- **Storage**: JSON Lines format for simple, portable data storage
- **QR Codes**: Generate HTTP endpoint links for phone-based logging

## Configuration

Environment variables:
- `SOURDOUGH_PORT` - Server port (default: 8080)
- `SOURDOUGH_DATA_DIR` - Data directory (default: ./data)
- `SOURDOUGH_SERVER_URL` - Server URL for CLI (default: http://localhost:8080)
