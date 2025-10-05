# Streamable HTTP MCP Server Example

Demonstrates a production-ready MCP server using Streamable HTTP transport with full CORS support.

## Quick Start

```bash
# Run with default settings (allows all origins)
go run main.go

# Run with specific allowed origin (for MCP Inspector)
go run main.go -origin "http://localhost:6274"

# Run on custom port
go run main.go -addr ":9000"
```

## Testing with MCP Inspector

1. Start the server:
```bash
go run main.go
```

2. Open MCP Inspector at http://localhost:6274

3. Connect to your server:
   - **URL**: `http://localhost:8080`
   - **Transport**: HTTP (Streamable)

## Testing with curl

### Test notification (no response expected):
```bash
curl -v -X POST http://localhost:8080 \
  -H "Origin: http://localhost:6274" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"notifications/initialized"}'
```

Expected response:
```
HTTP/1.1 202 Accepted
Access-Control-Allow-Origin: http://localhost:6274
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, Mcp-Session-Id, X-API-Key, Authorization, Last-Event-ID
Access-Control-Allow-Credentials: true
Access-Control-Expose-Headers: Mcp-Session-Id
```

### Test regular request:
```bash
curl -v -X POST http://localhost:8080 \
  -H "Origin: http://localhost:6274" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'
```

Expected response:
```
HTTP/1.1 200 OK
Access-Control-Allow-Origin: http://localhost:6274
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, Mcp-Session-Id, X-API-Key, Authorization, Last-Event-ID
Access-Control-Allow-Credentials: true
Access-Control-Expose-Headers: Mcp-Session-Id
Content-Type: application/json

{"jsonrpc":"2.0","result":{"tools":[...]},"id":1}
```

## Available Features

### Tools
- `add` - Add two numbers
- `multiply` - Multiply two numbers
- `greet` - Generate multilingual greetings (en, es, fr, de)

### Resources
- `config://server` - Server configuration (JSON)
- `status://health` - Server health status (JSON)
- `file:///{path}` - File reader template

### Prompts
- `greeting` - Generate greeting messages (name, language args)
- `calculation` - Create calculation requests (operation, a, b args)

## CORS Configuration

The server automatically handles CORS based on the `-origin` flag:

- **No origin flag** or **`-origin "*"`**: Allows any origin with credentials
- **`-origin "http://localhost:6274"`**: Only allows specified origin
- **`-origin "*.example.com"`**: Allows wildcard subdomains

CORS headers are set on:
- OPTIONS preflight requests
- POST request responses (both 200 and 202)
- GET SSE stream connections

## Troubleshooting

### CORS Error in Browser

If you see CORS errors:

1. Check the browser console for the exact error
2. Verify the Origin header matches your `-origin` flag
3. Check server logs for request details
4. Try with wildcard: `go run main.go -origin "*"`

### MCP Inspector Can't Connect

1. Ensure server is running: `curl http://localhost:8080`
2. Check CORS headers: `curl -v -H "Origin: http://localhost:6274" http://localhost:8080`
3. Try explicit origin: `go run main.go -origin "http://localhost:6274"`
4. Check MCP Inspector console for error details

### Notifications Failing

If initialize works but notifications fail:

1. Check server logs for the notification request
2. Verify 202 Accepted response is returned
3. Confirm CORS headers are present in 202 response
4. Check browser Network tab for response headers

## Architecture

```
Browser/Inspector (http://localhost:6274)
    ↓ POST with Origin header
streamhttp.Server.ServeHTTP()
    ↓ Validates origin
    ↓ Sets CORS headers
    ↓ Calls handlePOST()
handlePOST()
    ↓ Manages session
    ↓ Delegates to createMCPHandler()
createMCPHandler()
    ↓ Processes MCP message
    ↓ Returns 202 (notification) or 200 (request)
    ← Response with CORS headers preserved
```

## Key Implementation Details

1. **CORS headers set BEFORE handler**: Ensures headers are present on all responses
2. **Origin-specific credentials**: Uses request origin instead of `*` to allow credentials
3. **Expose-Headers**: Exposes `Mcp-Session-Id` for client access
4. **No header overwriting**: Handler doesn't set CORS headers (server handles it)
