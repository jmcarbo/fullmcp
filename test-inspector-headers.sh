#!/bin/bash

echo "Testing with MCP Inspector Headers (Referer instead of Origin)"
echo "=============================================================="
echo ""

echo "1. Testing notification (202 Accepted):"
echo "----------------------------------------"
curl -s -D - -X POST http://localhost:8080 \
  -H "Accept: application/json, text/event-stream" \
  -H "Content-Type: application/json" \
  -H "Mcp-Protocol-Version: 2025-06-18" \
  -H "Referer: http://localhost:6274/" \
  -d '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  -o /dev/null | grep -E "(HTTP|Access-Control)"

echo ""
echo "2. Testing initialize (200 OK):"
echo "--------------------------------"
curl -s -D - -X POST http://localhost:8080 \
  -H "Accept: application/json, text/event-stream" \
  -H "Content-Type: application/json" \
  -H "Mcp-Protocol-Version: 2025-06-18" \
  -H "Referer: http://localhost:6274/" \
  -d '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}},"id":1}' \
  -o /dev/null | grep -E "(HTTP|Access-Control)"

echo ""
echo "Expected Results:"
echo "================="
echo "Both should return:"
echo "  Access-Control-Allow-Origin: http://localhost:6274"
echo "  Access-Control-Allow-Credentials: true"
echo "  Access-Control-Allow-Headers: (includes Mcp-Protocol-Version)"
