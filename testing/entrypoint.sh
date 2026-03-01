#!/bin/sh
set -e

# Start the mock metadata server in the background
python3 /app/mock_metadata.py &
MOCK_PID=$!

# Wait for server to start
sleep 2

# Verify DNS override (optional, but good for debugging)
echo "Testing DNS resolution for metadata.google.internal..." >&2
ping -c 1 metadata.google.internal >&2 || echo "WARNING: DNS override might not work!" >&2

# Run the tool
echo "Running get_gcp_token..." >&2
export AWS_WEB_IDENTITY_TOKEN_FILE=/tmp/token
if get_gcp_token; then
    echo "SUCCESS: Tool executed successfully." >&2
    if [ "$(cat /tmp/token)" = "mock-oidc-token-for-testing" ]; then
        echo "VERIFICATION PASSED: Token content matches mock data." >&2
        kill $MOCK_PID
        exit 0
    else
        echo "VERIFICATION FAILED: Token content mismatch." >&2
        cat /tmp/token >&2
        kill $MOCK_PID
        exit 1
    fi
else
    echo "FAILURE: Tool execution failed." >&2
    kill $MOCK_PID
    exit 1
fi
