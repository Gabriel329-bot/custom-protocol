#!/bin/sh

set -e

CONFIG_FILE="/app/config.json"
ENV_FILE="/app/env"

echo "Custom Protocol Container starting..."

parse_env() {
    if [ -f "$ENV_FILE" ]; then
        set -a
        . "$ENV_FILE"
        set +a
    fi
}

get_config() {
    SERVER_PORT="${SERVER_PORT:-443}"
    SERVER_NAME="${SERVER_NAME:-www.microsoft.com}"
    FINGERPRINT="${FINGERPRINT:-Chrome}"
    OBFUSCATION="${OBFUSCATION:-3}"
    OBFS_PORT="${OBFS_PORT:-51820}"
    
    echo "Config:"
    echo "  Server Port: $SERVER_PORT"
    echo "  Server Name (SNI): $SERVER_NAME"
    echo "  Fingerprint: $FINGERPRINT"
    echo "  Obfuscation: $OBFUSCATION"
}

generate_keys() {
    if [ ! -f "/app/keys.json" ]; then
        echo "Generating keys..."
        PRIVATE_KEY=$(customproto keygen 2>/dev/null | grep "Private Key" | cut -d: -f2 | tr -d ' ')
        PUBLIC_KEY=$(customproto keygen 2>/dev/null | grep "Public Key" | cut -d: -f2 | tr -d ' ')
        
        cat > /app/keys.json << EOF
{
    "private_key": "$PRIVATE_KEY",
    "public_key": "$PUBLIC_KEY"
}
EOF
        echo "Keys generated!"
    fi
}

start_server() {
    echo "Starting custom protocol server..."
    
    exec customproto server -l ":$SERVER_PORT" \
        --server-name "$SERVER_NAME" \
        --fingerprint "$FINGERPRINT" \
        --obfuscation "$OBFUSCATION"
}

start_client() {
    SERVER_HOST="${SERVER_HOST:-server}"
    SERVER_ADDR="${SERVER_HOST}:${SERVER_PORT:-443}"
    
    echo "Starting custom protocol client..."
    echo "Connecting to: $SERVER_ADDR"
    
    exec customproto client "$SERVER_ADDR" \
        --server-name "$SERVER_NAME" \
        --fingerprint "$FINGERPRINT" \
        --obfuscation "$OBFUSCATION"
}

parse_env
get_config

MODE="${MODE:-server}"

case "$MODE" in
    server)
        start_server
        ;;
    client)
        start_client
        ;;
    *)
        echo "Unknown mode: $MODE"
        echo "Usage: MODE=server|client"
        exit 1
        ;;
esac