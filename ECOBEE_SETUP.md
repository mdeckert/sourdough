# Ecobee Integration Setup Guide

This guide explains how to set up automatic kitchen temperature logging from your Ecobee thermostat.

## Overview

The sourdough logger can automatically fetch kitchen temperature from your Ecobee thermostat and include it with every event you log (mixing, folding, shaping, etc.). This eliminates manual temperature entry and provides more accurate tracking.

## Prerequisites

- Ecobee thermostat already set up in Apple HomeKit
- iPhone with Home app configured
- Linux server running 24/7 (192.168.1.50 in this setup)
- Node.js and npm installed

## Installation Steps

### 1. Install Homebridge

Run the provided setup script:

```bash
cd /home/mdeckert/sourdough
./scripts/setup_homebridge.sh
```

This script will:
- Install Node.js (if not already installed)
- Install homebridge and homebridge-config-ui-x
- Install homebridge-http-webhooks plugin
- Create systemd service for homebridge
- Start homebridge automatically

### 2. Configure Homebridge

1. **Access Homebridge UI**:
   ```
   http://192.168.1.50:8581
   ```

2. **Login**:
   - First time: Create admin username/password
   - Subsequent: Use your credentials

3. **Pair with HomeKit**:
   - Open Home app on iPhone
   - Tap "Add Accessory"
   - Scan QR code shown in homebridge UI
   - Or manually enter PIN: `031-45-154`

4. **Wait for Accessories to Load**:
   - Homebridge will discover all HomeKit accessories
   - Your Ecobee should appear in the list
   - Note the **exact name** of the temperature sensor

### 3. Install HTTP Webhooks Plugin

If not already installed:

```bash
sudo npm install -g homebridge-http-webhooks
```

Add to homebridge config via UI:
1. Go to Plugins → homebridge-http-webhooks
2. Click "Settings"
3. Configure webhook port: `51828`
4. Save and restart homebridge

### 4. Configure Sourdough Server

Add environment variables to the systemd service:

```bash
sudo nano /etc/systemd/system/sourdough.service
```

Add these lines in the `[Service]` section:

```ini
Environment="ECOBEE_URL=http://localhost:51828"
Environment="ECOBEE_DEVICE=Kitchen"
```

**Note**: Replace `Kitchen` with the exact name of your Ecobee sensor as it appears in homebridge.

Reload and restart:

```bash
sudo systemctl daemon-reload
sudo systemctl restart sourdough
```

### 5. Verify Integration

Check the sourdough logs:

```bash
sudo journalctl -u sourdough -n 50
```

You should see:
```
Ecobee integration enabled: http://localhost:51828/Kitchen
```

If you see:
```
Ecobee integration disabled (set ECOBEE_URL and ECOBEE_DEVICE to enable)
```

The environment variables aren't set correctly.

### 6. Test Temperature Fetch

Log any event via QR code or API:

```bash
curl -X POST http://192.168.1.50:8080/log/mixed
```

Check the logs:

```bash
sudo journalctl -u sourdough -f
```

You should see:
```
Auto-fetched kitchen temp from Ecobee: 72.5°F
```

View the event:

```bash
curl http://192.168.1.50:8080/status | grep -i temp
```

The temperature should be included automatically.

## How It Works

1. **Event Logging**: When you scan a QR code or log an event
2. **Auto-Fetch**: Server checks if Ecobee integration is enabled
3. **Temperature Request**: Fetches current temperature from homebridge
4. **Temperature Added**: Automatically adds kitchen temp to the event
5. **Saved**: Event saved with both your action AND ambient temperature

## Configuration

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `ECOBEE_URL` | Homebridge webhook base URL | `http://localhost:51828` |
| `ECOBEE_DEVICE` | Ecobee sensor name in homebridge | `Kitchen` or `Ecobee Thermostat` |

### When Temperature is Auto-Fetched

Temperature is automatically fetched for:
- ✅ All workflow events (mixed, fold, shaped, etc.)
- ✅ Note logging
- ✅ Loaf completion
- ❌ NOT for manual temperature logging (would override your value)

### Disabling Auto-Fetch

To disable temporarily:

```bash
sudo nano /etc/systemd/system/sourdough.service
# Comment out or remove ECOBEE_* lines
sudo systemctl daemon-reload
sudo systemctl restart sourdough
```

## Troubleshooting

### Homebridge Not Starting

```bash
sudo systemctl status homebridge
sudo journalctl -u homebridge -n 50
```

Common issues:
- Port 51826 already in use
- Permission issues with ~/.homebridge directory
- Node.js version too old

### Ecobee Not Appearing

- Verify Ecobee is in HomeKit on your iPhone
- Check homebridge logs for pairing errors
- Try removing and re-adding homebridge in Home app

### Temperature Not Auto-Fetching

Check sourdough logs:

```bash
sudo journalctl -u sourdough -f
# Then log an event
curl -X POST http://192.168.1.50:8080/log/mixed
```

Look for:
- "Auto-fetched kitchen temp..." = Working!
- "Warning: Failed to fetch..." = Connection issue
- No message = Integration disabled

Verify homebridge is running:

```bash
sudo systemctl status homebridge
curl http://localhost:51828/Kitchen
```

### Wrong Temperature Unit

Homebridge returns Celsius. The sourdough server automatically converts to Fahrenheit. If temperatures seem wrong, check the conversion in logs.

### Device Name Mismatch

Find exact device name:

```bash
# List all homebridge accessories
curl http://localhost:51828/
```

Or check homebridge UI → Accessories tab

## Advanced Configuration

### Custom Homebridge Port

If you changed the webhook port:

```bash
# Edit sourdough service
sudo nano /etc/systemd/system/sourdough.service

# Update ECOBEE_URL
Environment="ECOBEE_URL=http://localhost:YOUR_PORT"
```

### Remote Homebridge

If homebridge runs on a different machine:

```bash
Environment="ECOBEE_URL=http://192.168.1.100:51828"
```

### Multiple Sensors

Currently, only one sensor is supported. To use a different sensor, change `ECOBEE_DEVICE` to its name.

## Maintenance

### Update Homebridge

```bash
sudo npm update -g homebridge homebridge-config-ui-x homebridge-http-webhooks
sudo systemctl restart homebridge
```

### View Homebridge Logs

```bash
sudo journalctl -u homebridge -f
```

### Backup Homebridge Config

```bash
cp ~/.homebridge/config.json ~/.homebridge/config.json.backup
```

## Benefits

- **Automatic**: No manual temperature entry needed
- **Accurate**: Uses actual ambient temperature, not estimates
- **Consistent**: Every event has temperature data
- **Convenient**: Just scan QR codes as normal
- **Historical**: Complete temperature timeline throughout bake

## Support

If you encounter issues:

1. Check homebridge is running: `sudo systemctl status homebridge`
2. Check sourdough logs: `sudo journalctl -u sourdough -n 50`
3. Test homebridge endpoint: `curl http://localhost:51828/YOUR_DEVICE`
4. Verify environment variables: `systemctl cat sourdough | grep ECOBEE`
