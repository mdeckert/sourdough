# Quick Start Guide

## Daily Workflow

### Starting a Bake

```bash
# Take starter out of freezer
sourdough start

# Or scan "Starter Out" QR code
```

### Logging Events Throughout the Day

```bash
# Feed starter
sourdough log fed

# Levain is ready
sourdough log levain-ready

# Mix dough (start fermentolyse)
sourdough log mixed

# Each fold
sourdough log fold  # auto-increments count

# Shape the dough
sourdough log shaped

# Put in fridge
sourdough log fridge-in

# Take out of fridge (next day)
sourdough log fridge-out

# Put in oven
sourdough log oven-in
```

### Logging Temperatures

```bash
# Kitchen temperature
sourdough temp 76

# Or scan temperature QR code
```

### Check Status Anytime

```bash
sourdough status
```

Output shows timeline with durations:
```
Bake Status - 2025-10-07
==================================================
08:30  starter-out     [68°F]
12:30  fed             (+4h)
18:00  levain-ready    (+5h30m)
18:15  mixed           [76°F] (+15m)
19:15  fold            #1 (+1h)
20:15  fold            #2 (+1h)
21:15  fold            #3 (+1h)
22:30  shaped          (+1h15m)
22:45  fridge-in       (+15m)
--------------------------------------------------
Total elapsed: 14h15m
```

### Complete the Bake

```bash
sourdough complete
```

This will ask you to assess:
- Proof level (underproofed/good/overproofed)
- Crumb quality (1-10)
- Browning (none/slight/good/over)
- Overall score (1-10)
- Notes (optional)

### Review Past Bakes

```bash
# List recent bakes
sourdough history

# Review specific bake
sourdough review 2025-10-07
```

## QR Code Workflow

1. Print `qrcodes/sheet.png`
2. Cut out and label QR codes
3. Stick on fridge
4. Scan with phone to log events

**No need to unlock phone or open app!** Just scan and tap notification.

## Common Event Sequence

```
starter-out  →  fed  →  levain-ready  →  mixed
   ↓
fold (repeat 3-4x every 30-60min)
   ↓
shaped  →  fridge-in
   ↓
[overnight in fridge]
   ↓
fridge-out  →  oven-in  →  bake-complete
```

## Tips

- **Log temperatures frequently**: kitchen temp affects bulk fermentation timing
- **Use QR codes**: Fastest way to log while hands are floury
- **Check status often**: See how timing compares to previous bakes
- **Complete assessment**: This is key for learning what works
- **Note patterns**: If a bake goes well, check the timeline and temps

## Timing Reference (Your Recipe)

Based on your goal to fix underproofing:

**Target bulk fermentation**: 4-6 hours at 74-78°F
- Watch for 50-75% volume increase
- Dough should be puffy, jiggly, slightly domed

**Cold proof**: 12-18 hours
- Longer is generally better for flavor
- Helps prevent overbrowning issues

**Signs of proper proof**:
- Dough holds shape when turned out
- Springs back slowly when poked
- Shows some air bubbles under surface

The system will help you correlate temps with outcomes!
