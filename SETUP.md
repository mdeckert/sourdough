# Sourdough System Setup Guide

## Prerequisites

This system requires Go 1.21 or later. Follow these steps to get everything running.

## Step 1: Install Go

```bash
# Download Go 1.21 (or later)
cd ~
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz

# Remove any previous Go installation and extract
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

## Step 2: Build the System

```bash
cd /home/mdeckert/sourdough

# Download dependencies
go mod download
go mod tidy

# Build all components
make build

# Or build individually:
# make server  # builds bin/sourdough-server
# make cli     # builds bin/sourdough
# make qrgen   # builds bin/qrgen
```

## Step 3: Test Locally

```bash
# Start server in foreground (for testing)
./bin/sourdough-server

# In another SSH session, test the CLI:
./bin/sourdough start
./bin/sourdough log mixed
./bin/sourdough temp 76
./bin/sourdough status
```

## Step 4: Install as System Service

```bash
# Install and start the service
make install-service

# Check service status
sudo systemctl status sourdough

# View logs
sudo journalctl -u sourdough -f

# Stop/start/restart service
sudo systemctl stop sourdough
sudo systemctl start sourdough
sudo systemctl restart sourdough
```

## Step 5: Install CLI Tool

```bash
# Install CLI to system PATH
make install

# Now you can use from anywhere:
sourdough start
sourdough status
```

## Step 6: Generate QR Codes

```bash
# Find your server's local IP address
ip addr show | grep "inet " | grep -v 127.0.0.1

# Generate QR codes (replace with your server's IP)
./bin/qrgen http://192.168.1.100:8080

# QR codes will be in ./qrcodes/
# Print qrcodes/sheet.png and cut out individual codes
```

## Step 7: Test from Phone

1. Make sure your phone is on the same network as the server
2. Scan one of the QR codes with your phone's camera
3. Tap the notification to open the URL
4. Event should be logged!
5. Verify with: `sourdough status`

## Configuration

### Environment Variables

Create a file `/etc/systemd/system/sourdough.service.d/override.conf`:

```ini
[Service]
Environment="SOURDOUGH_PORT=8080"
Environment="SOURDOUGH_DATA_DIR=/home/mdeckert/sourdough/data"
```

Then reload: `sudo systemctl daemon-reload && sudo systemctl restart sourdough`

### For CLI Tool

Add to `~/.bashrc`:

```bash
export SOURDOUGH_SERVER_URL="http://localhost:8080"
export SOURDOUGH_DATA_DIR="/home/mdeckert/sourdough/data"
```

## Troubleshooting

### Server won't start
```bash
# Check if port 8080 is in use
sudo netstat -tlnp | grep 8080

# Check service logs
sudo journalctl -u sourdough -n 50
```

### CLI can't connect to server
```bash
# Make sure server is running
sudo systemctl status sourdough

# Test server directly
curl http://localhost:8080/health

# Check firewall (if accessing from phone)
sudo ufw status
sudo ufw allow 8080/tcp  # if needed
```

### QR codes don't work from phone
- Ensure phone and server are on same network
- Verify server IP with `ip addr`
- Test URL in phone browser first: `http://YOUR_IP:8080/health`
- Check firewall allows connections on port 8080

## Data Backup

Your bake data is stored in `./data/` as JSON Lines files:

```bash
# Backup all bakes
tar -czf sourdough-backup-$(date +%Y%m%d).tar.gz data/

# Restore
tar -xzf sourdough-backup-YYYYMMDD.tar.gz
```

## Uninstall

```bash
# Stop and remove service
make uninstall-service

# Remove CLI from system path
sudo rm /usr/local/bin/sourdough

# Remove everything
cd ~
rm -rf sourdough/
```

## Next Steps

After collecting 3-4 bakes of data:
- Review patterns with `sourdough history`
- Analyze timing vs temperature correlations
- Adjust your process based on data
- Future enhancement: automated analysis and timing predictions
