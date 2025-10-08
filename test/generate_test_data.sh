#!/bin/bash
# Generate comprehensive test dataset for sourdough logger
# This creates a 2-day bake with all event types for testing

DATA_DIR="./data"
TEST_FILE="$DATA_DIR/bake_2025-10-05_08-00.jsonl"

echo "Generating comprehensive test dataset..."

# Create data directory if it doesn't exist
mkdir -p "$DATA_DIR"

# Generate the test JSONL file
cat > "$TEST_FILE" << 'EOF'
{"timestamp":"2025-10-05T08:00:00-07:00","event":"starter-out"}
{"timestamp":"2025-10-05T08:05:00-07:00","event":"fed"}
{"timestamp":"2025-10-05T08:10:00-07:00","event":"temperature","temp_f":68}
{"timestamp":"2025-10-05T08:15:00-07:00","event":"note","note":"Starter looking very bubbly and active"}
{"timestamp":"2025-10-05T13:00:00-07:00","event":"levain-ready"}
{"timestamp":"2025-10-05T13:05:00-07:00","event":"temperature","temp_f":72}
{"timestamp":"2025-10-05T13:15:00-07:00","event":"mixed"}
{"timestamp":"2025-10-05T13:20:00-07:00","event":"temperature","dough_temp_f":76}
{"timestamp":"2025-10-05T13:25:00-07:00","event":"note","note":"Dough feels smooth and well-hydrated, slightly sticky"}
{"timestamp":"2025-10-05T14:30:00-07:00","event":"fold","fold_count":1}
{"timestamp":"2025-10-05T14:35:00-07:00","event":"temperature","dough_temp_f":74}
{"timestamp":"2025-10-05T15:30:00-07:00","event":"fold","fold_count":2}
{"timestamp":"2025-10-05T16:30:00-07:00","event":"fold","fold_count":3}
{"timestamp":"2025-10-05T16:35:00-07:00","event":"temperature","temp_f":73}
{"timestamp":"2025-10-05T16:35:30-07:00","event":"temperature","dough_temp_f":75}
{"timestamp":"2025-10-05T17:30:00-07:00","event":"fold","fold_count":4}
{"timestamp":"2025-10-05T17:35:00-07:00","event":"note","note":"Great dough strength, holds shape well, passing windowpane test"}
{"timestamp":"2025-10-05T19:00:00-07:00","event":"shaped"}
{"timestamp":"2025-10-05T19:05:00-07:00","event":"temperature","dough_temp_f":76}
{"timestamp":"2025-10-05T19:30:00-07:00","event":"fridge-in"}
{"timestamp":"2025-10-05T19:35:00-07:00","event":"note","note":"Shaped into batard, nice tight skin, planning 14 hour cold proof"}
{"timestamp":"2025-10-06T09:35:00-07:00","event":"note","note":"Good rise overnight, nice wobble when shaken, looks ready"}
{"timestamp":"2025-10-06T10:00:00-07:00","event":"note","note":"Starting oven preheat to 500°F with Dutch oven inside"}
{"timestamp":"2025-10-06T10:30:00-07:00","event":"oven-in"}
{"timestamp":"2025-10-06T10:31:00-07:00","event":"temperature","temp_f":500}
{"timestamp":"2025-10-06T10:50:00-07:00","event":"temperature","temp_f":450}
{"timestamp":"2025-10-06T10:50:30-07:00","event":"note","note":"Removed lid, good oven spring visible"}
{"timestamp":"2025-10-06T11:00:00-07:00","event":"temperature","dough_temp_f":180}
{"timestamp":"2025-10-06T11:05:00-07:00","event":"temperature","dough_temp_f":200}
{"timestamp":"2025-10-06T11:10:00-07:00","event":"temperature","dough_temp_f":208}
{"timestamp":"2025-10-06T11:10:30-07:00","event":"oven-out"}
{"timestamp":"2025-10-06T11:15:00-07:00","event":"note","note":"Beautiful golden-brown color, hollow sound when tapped, excellent ear on the score"}
{"timestamp":"2025-10-06T14:00:00-07:00","event":"loaf-complete","data":{"assessment":{"proof_level":"good","crumb_quality":9,"browning":"good","score":9,"notes":"Excellent open crumb with nice irregular holes, good gluten development. Crust is perfectly crispy. Flavor is complex with nice tang. One of the best bakes yet!"}}}
EOF

echo "✓ Test dataset created: $TEST_FILE"
echo ""
echo "Dataset details:"
echo "  - Date: 2025-10-05 (Oct 5-6, 2025)"
echo "  - Duration: ~30 hours (2 days)"
echo "  - Events: 33 events"
echo "  - Features: All event types, temps (kitchen/dough/oven/loaf), notes, folds, assessment"
echo ""
echo "View at: http://192.168.1.50:8080/view/status?date=2025-10-05_08-00"
