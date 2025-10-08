# QR Code Generation Guide

## Quick Start

Generate QR codes with your server's IP address:

```bash
./bin/qrgen http://192.168.1.50:8080
```

Test that they work:

```bash
make test-qr
```

Print the PDF or sheet:
- PDF: `qrcodes/qrcodes.pdf`
- Print-ready sheet: `qrcodes/sheet.png`

## Important Notes

### Always Use IP Address, Not Localhost

❌ **WRONG:**
```bash
./bin/qrgen http://localhost:8080
```

✅ **CORRECT:**
```bash
./bin/qrgen http://192.168.1.50:8080
```

**Why?** QR codes are scanned from your phone, which needs to access the server over the network. Localhost only works from the server itself.

### Testing QR Codes

**Always test after generation:**

```bash
# Quick test
make test-qr

# Or manually test one endpoint
curl -X POST http://192.168.1.50:8080/log/mixed
```

## Validation

The qrgen tool now validates URLs:

1. **Rejects help flags**: `./bin/qrgen --help` won't generate bad QR codes
2. **Rejects localhost**: `./bin/qrgen http://localhost:8080` will error with helpful message
3. **Requires IP address**: Forces you to use a reachable URL

## Testing

Three levels of QR code testing:

1. **URL validation**: Ensures invalid URLs are rejected
2. **File generation**: Verifies all QR code files are created
3. **Endpoint testing**: Tests that QR code URLs actually work

Run with:
```bash
make test-qr
```

**Note:** The test creates real data by calling the API, but it automatically backs up and restores your data directory, so your current loaf won't be polluted.

## What Went Wrong Before

**Problem:** The qrgen tool would accept any argument, including `--help`, and generate QR codes with invalid URLs.

**Root cause:** No URL validation and no automated testing after generation.

**Solution:**
1. Added URL validation to reject `--help`, `-h`, and `localhost` URLs
2. Created automated test suite (`test/test_qr.sh`)
3. Added `make test-qr` target
4. Included QR tests in `make test-all`

## Regenerating QR Codes

If you change your server's IP address:

```bash
# Build the QR generator
make qrgen

# Generate new QR codes
./bin/qrgen http://YOUR_NEW_IP:8080

# Test them
make test-qr

# Print new sheet
# Print qrcodes/qrcodes.pdf or qrcodes/sheet.png
```

## Troubleshooting

### QR codes point to wrong URL

**Check current QR codes:**
```bash
# The test will show what URL they're using
make test-qr
```

**Regenerate:**
```bash
./bin/qrgen http://192.168.1.50:8080
```

### QR code doesn't work when scanned

1. **Check server is running:**
   ```bash
   sudo systemctl status sourdough
   ```

2. **Check server is accessible from phone:**
   ```bash
   curl http://192.168.1.50:8080/health
   ```

3. **Test the specific endpoint:**
   ```bash
   curl -X POST http://192.168.1.50:8080/log/mixed
   ```

4. **Check firewall** (if curl works from server but not phone):
   ```bash
   sudo ufw status
   # Port 8080 should be allowed
   ```

## Best Practices

1. ✅ **Always use IP address** when generating QR codes
2. ✅ **Run `make test-qr`** after generating QR codes
3. ✅ **Test with your phone** before printing
4. ✅ **Keep QR codes in version control** (the .png and .pdf files)
5. ✅ **Regenerate if IP changes** (and test again!)

## Files Generated

```
qrcodes/
├── qrcodes.pdf          # Full PDF with all QR codes
├── sheet.png            # Print-ready sheet (all codes on one page)
├── start.png            # Individual QR code files
├── fed.png
├── mixed.png
├── fold.png
├── shaped.png
├── fridge-in.png
├── oven-in.png
├── oven-out.png
├── temp.png
├── notes.png
├── complete.png
├── status.png
├── history.png
└── qr-pdf.png
```

## Integration with Ecobee

With Ecobee integration enabled (via Home Assistant), QR codes automatically include kitchen temperature:

```bash
# Scan "Mixed" QR code → logs mixing event with current temp
# Result: {"event":"mixed","temp_f":70.88, ...}
```

No manual temperature entry needed!
