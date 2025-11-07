#!/bin/bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT="$ROOT_DIR/cmd/kaunta/assets/kaunta.min.js"
SOURCE="$ROOT_DIR/tracker/kaunta.js"

echo "🔨 Building Kaunta tracker with Bun..."
echo ""

if ! command -v bun &> /dev/null; then
    echo "❌ Bun not found. Install it from https://bun.sh before running this script."
    exit 1
fi

pushd "$ROOT_DIR" > /dev/null
bun run build:tracker
popd > /dev/null

echo "✅ Tracker bundle created at $OUTPUT"

# Generate SRI hash
echo ""
echo "🔐 SRI Hash (sha384):"
echo "   sha384-$(openssl dgst -sha384 -binary "$OUTPUT" | openssl base64 -A)"

# Gzip for size testing
gzip -c "$OUTPUT" > "$OUTPUT.gz"

# Report sizes
echo ""
echo "📊 Sizes:"
printf "   Original: %'8d bytes (%.2f KB)\n" $(wc -c < "$SOURCE") $(echo "scale=2; $(wc -c < "$SOURCE")/1024" | bc)
printf "   Minified: %'8d bytes (%.2f KB)\n" $(wc -c < "$OUTPUT") $(echo "scale=2; $(wc -c < "$OUTPUT")/1024" | bc)
printf "   Gzipped:  %'8d bytes (%.2f KB)\n" $(wc -c < "$OUTPUT.gz") $(echo "scale=2; $(wc -c < "$OUTPUT.gz")/1024" | bc)

# Calculate savings
original_size=$(wc -c < "$SOURCE")
minified_size=$(wc -c < "$OUTPUT")
gzipped_size=$(wc -c < "$OUTPUT.gz")

minified_percent=$(echo "scale=1; 100 - ($minified_size * 100 / $original_size)" | bc)
gzipped_percent=$(echo "scale=1; 100 - ($gzipped_size * 100 / $original_size)" | bc)

echo ""
echo "💾 Savings:"
echo "   Minified: -${minified_percent}%"
echo "   Gzipped:  -${gzipped_percent}%"

# Cleanup
rm "$OUTPUT.gz"

echo ""
echo "✨ Done! Output: $OUTPUT"
