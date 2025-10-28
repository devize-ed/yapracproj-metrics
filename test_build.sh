#!/bin/bash

# Exit immediately if a command fails
set -e

# test values
TEST_VERSION="v1.2.3-test"
TEST_DATE="2025-01-01T12:00:00Z"
TEST_COMMIT="testcommit123"

# temporary binary file name
BINARY_NAME="./test"

# build application with ldflags
echo "Building application with ldflags..."
go build -o $BINARY_NAME \
    -ldflags "\
    -X 'main.buildVersion=${TEST_VERSION}' \
    -X 'main.buildDate=${TEST_DATE}' \
    -X 'main.buildCommit=${TEST_COMMIT}'" \
    ./cmd/server


# execute binary in background and capture output to temp file
TEMP_OUTPUT=$(mktemp)
./$BINARY_NAME > "$TEMP_OUTPUT" 2>&1 &
BINARY_PID=$!

# wait a moment for the binary to print version info
sleep 0.5

# kill the binary process
kill $BINARY_PID 2>/dev/null || true
wait $BINARY_PID 2>/dev/null || true

# read captured output
OUTPUT=$(cat "$TEMP_OUTPUT")
rm "$TEMP_OUTPUT"

# clean up temporary binary file
rm $BINARY_NAME

# check output version
if ! echo "$OUTPUT" | grep -q "$TEST_VERSION"; then
    echo "❌ Build version not found"
    exit 1
fi

# check output date
if ! echo "$OUTPUT" | grep -q "$TEST_DATE"; then
    echo "❌ Build date not found"
    exit 1
fi

# check output commit
if ! echo "$OUTPUT" | grep -q "$TEST_COMMIT"; then
    echo "❌ Build commit not found"
    exit 1
fi
