#!/bin/bash
#
# Overnight fuzz testing script for msgpck
#
# Usage:
#   ./scripts/fuzz-overnight.sh              # Run all fuzz tests (~40min each, ~8h total)
#   ./scripts/fuzz-overnight.sh 1h           # Run each test for 1 hour
#   ./scripts/fuzz-overnight.sh 10m FuzzDecoder  # Run specific test for 10 minutes
#   FUZZ_PARALLEL=8 ./scripts/fuzz-overnight.sh  # Use 8 workers (default: 4)
#
# The script will:
#   - Run all fuzz targets (or a specific one if specified)
#   - Save output to logs/fuzz-YYYY-MM-DD/
#   - Continue to next test if one finishes or crashes
#   - Print a summary at the end
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

# Default time per fuzz target (8 hours / 12 targets ≈ 40 minutes each)
TIME_PER_TARGET="${1:-40m}"
SPECIFIC_TARGET="${2:-}"

# Limit parallel workers to avoid memory exhaustion (default: 4)
FUZZ_PARALLEL="${FUZZ_PARALLEL:-4}"

# Create log directory
DATE=$(date +%Y-%m-%d-%H%M%S)
LOG_DIR="$PROJECT_ROOT/logs/fuzz-$DATE"
mkdir -p "$LOG_DIR"

# All fuzz targets (12 total)
TARGETS="FuzzDecoder FuzzRoundTrip FuzzRoundTripInt FuzzRoundTripFloat FuzzMapDecode FuzzMapStringString FuzzStructDecode FuzzCachedStructDecoder FuzzDecodeStructFunc FuzzStructEncoder FuzzNestedStructures FuzzLargeCollections"

# Track results
RESULTS_FILE="$LOG_DIR/results.txt"
CRASHED_FILE="$LOG_DIR/crashed.txt"
touch "$RESULTS_FILE" "$CRASHED_FILE"
TOTAL_FINDINGS=0

copy_corpus_to_testdata() {
    echo ""
    echo "=========================================="
    echo "Copying fuzz corpus to testdata/"
    echo "=========================================="

    local GOCACHE_FUZZ="$(go env GOCACHE)/fuzz/github.com/freeeve/msgpck"
    local TESTDATA_FUZZ="$PROJECT_ROOT/testdata/fuzz"

    if [ -d "$GOCACHE_FUZZ" ]; then
        mkdir -p "$TESTDATA_FUZZ"
        local before_count=$(find "$TESTDATA_FUZZ" -type f 2>/dev/null | wc -l | tr -d ' ')
        cp -r "$GOCACHE_FUZZ"/* "$TESTDATA_FUZZ"/ 2>/dev/null || true
        local after_count=$(find "$TESTDATA_FUZZ" -type f | wc -l | tr -d ' ')
        local new_count=$((after_count - before_count))
        echo "Copied corpus from Go cache to testdata/fuzz/"
        echo "New files added: $new_count (total: $after_count)"
    else
        echo "No fuzz cache found at $GOCACHE_FUZZ"
    fi
    return 0
}

print_summary() {
    echo ""
    echo "=========================================="
    echo "FUZZ TESTING SUMMARY"
    echo "=========================================="
    echo "Log directory: $LOG_DIR"
    echo "Time per target: $TIME_PER_TARGET"
    echo ""

    if [ -s "$CRASHED_FILE" ]; then
        echo "CRASHED TARGETS:"
        while read -r target; do
            echo "   - $target (see $LOG_DIR/$target.log)"
        done < "$CRASHED_FILE"
        echo ""
    fi

    echo "Results:"
    if [ -s "$RESULTS_FILE" ]; then
        cat "$RESULTS_FILE" | while read -r line; do
            echo "   $line"
        done
    fi

    echo ""
    echo "Total new corpus entries: $TOTAL_FINDINGS"
    echo ""
    echo "To view logs:"
    echo "   ls -la $LOG_DIR/"
    echo ""
    echo "To check for failures:"
    echo "   grep -l 'FAIL\|panic\|crash' $LOG_DIR/*.log 2>/dev/null || echo 'No failures found'"
}

# Cleanup on exit
cleanup() {
    echo ""
    echo "=========================================="
    echo "Fuzz testing interrupted or completed"
    echo "=========================================="
    copy_corpus_to_testdata
    print_summary
    exit 0
}
trap cleanup SIGINT SIGTERM

run_fuzz_target() {
    local target="$1"
    local logfile="$LOG_DIR/$target.log"

    echo ""
    echo "=========================================="
    echo "Running $target for $TIME_PER_TARGET"
    echo "Log: $logfile"
    echo "Started: $(date)"
    echo "=========================================="

    # Run fuzz test, capture output
    # Use ^...$ anchors for exact match
    # Use -parallel to limit workers and avoid memory exhaustion
    set +e
    go test -fuzz="^${target}\$" -fuzztime="$TIME_PER_TARGET" -parallel="$FUZZ_PARALLEL" -v . 2>&1 | tee "$logfile"
    local exit_code=$?
    set -e

    # Check results
    if [ $exit_code -ne 0 ]; then
        echo "WARNING: $target exited with code $exit_code"
        echo "$target" >> "$CRASHED_FILE"
        echo "$target: CRASHED (exit $exit_code)" >> "$RESULTS_FILE"
    else
        # Count new corpus entries
        local new_entries
        new_entries=$(grep "new interesting" "$logfile" 2>/dev/null | wc -l | tr -d ' ')
        new_entries=${new_entries:-0}
        TOTAL_FINDINGS=$((TOTAL_FINDINGS + new_entries))
        echo "$target: OK (+$new_entries corpus entries)" >> "$RESULTS_FILE"
        echo "OK: $target completed (+$new_entries new entries)"
    fi

    echo "Finished: $(date)"
}

is_in_list() {
    local target="$1"
    local list="$2"
    for item in $list; do
        if [ "$item" = "$target" ]; then
            return 0
        fi
    done
    return 1
}

echo "=========================================="
echo "msgpck Overnight Fuzz Testing"
echo "=========================================="
echo "Started: $(date)"
echo "Time per target: $TIME_PER_TARGET"
echo "Parallel workers: $FUZZ_PARALLEL"
echo "Log directory: $LOG_DIR"
echo ""

if [ -n "$SPECIFIC_TARGET" ]; then
    # Run specific target
    if is_in_list "$SPECIFIC_TARGET" "$TARGETS"; then
        echo "Running specific target: $SPECIFIC_TARGET"
        run_fuzz_target "$SPECIFIC_TARGET"
    else
        echo "Unknown fuzz target: $SPECIFIC_TARGET"
        echo ""
        echo "Available targets:"
        for t in $TARGETS; do
            echo "  - $t"
        done
        exit 1
    fi
else
    # Count targets
    target_count=0
    for t in $TARGETS; do target_count=$((target_count + 1)); done

    echo "Running all $target_count fuzz targets"
    echo ""

    for target in $TARGETS; do
        run_fuzz_target "$target"
    done
fi

copy_corpus_to_testdata
print_summary
