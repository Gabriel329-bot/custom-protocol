#!/bin/bash

set -e

echo "=== Custom Protocol Docker Build ==="

echo "Building customproto for Linux..."
GOOS=linux GOARCH=amd64 go build -o customproto .
echo "Built: customproto"

echo "Building Docker image..."
docker build -t customproto:latest .
echo "Image built: customproto:latest"

echo ""
echo "=== Ready to use ==="
echo ""
echo "Server mode:"
echo "  docker run -d -p 443:443 -e MODE=server customproto:latest"
echo ""
echo "Client mode:"
echo "  docker run -it --rm customproto:latest client -connect server:443"
echo ""

echo "Or use docker-compose:"
echo "  docker-compose up -d"
echo ""

echo "=== Files for AmneziaVPN Custom Container ==="
echo ""
echo "Image: customproto:latest"
echo ""
echo "Environment variables:"
echo "  MODE=server|client"
echo "  SERVER_NAME=www.microsoft.com"
echo "  FINGERPRINT=Chrome"
echo "  OBFUSCATION=3"