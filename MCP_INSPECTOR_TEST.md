# Testing with MCP Inspector

## ✅ FIXED: Referer Header Support

The server now properly handles clients that send `Referer` instead of `Origin` headers (like MCP Inspector).

## Verified Working Configuration

The streamhttp server has been tested and **CORS headers are working correctly** for both regular requests and notifications, including with MCP Inspector's header format.

### Test Results

```bash
# Notification (no response body)
curl -H "Origin: http://localhost:6274" -X POST http://localhost:8080 \
  -d '{"jsonrpc":"2.0","method":"notifications/initialized"}'

Response:
  HTTP/1.1 202 Accepted
  Access-Control-Allow-Origin: http://localhost:6274
  Access-Control-Allow-Methods: GET, POST, OPTIONS
  Access-Control-Allow-Headers: Content-Type, Mcp-Session-Id, X-API-Key, Authorization, Last-Event-ID
  Access-Control-Allow-Credentials: true
  Access-Control-Expose-Headers: Mcp-Session-Id

# Initialize (with response)
curl -H "Origin: http://localhost:6274" -X POST http://localhost:8080 \
  -d '{"jsonrpc":"2.0","method":"initialize","params":{},"id":1}'

Response:
  HTTP/1.1 200 OK
  Access-Control-Allow-Origin: http://localhost:6274
  Access-Control-Allow-Methods: GET, POST, OPTIONS
  Access-Control-Allow-Headers: Content-Type, Mcp-Session-Id, X-API-Key, Authorization, Last-Event-ID
  Access-Control-Allow-Credentials: true
  Access-Control-Expose-Headers: Mcp-Session-Id
```

## Setup Instructions

### 1. Start the Server

```bash
cd examples/streamhttp
go run main.go
```

You should see:
```
═══════════════════════════════════════════════════════════
Starting Streamable HTTP MCP Server
═══════════════════════════════════════════════════════════
Address:         :8080
Transport:       Streamable HTTP (MCP 2025-03-26)
CORS Origin:     *

For MCP Inspector, connect to: http://localhost:8080
```

### 2. Open MCP Inspector

Visit: http://localhost:6274

### 3. Connect to Server

In the MCP Inspector interface:
1. Click "Add Server" or "Connect"
2. Enter server URL: `http://localhost:8080`
3. Select transport: "HTTP" or "Streamable HTTP"
4. Click "Connect"

### 4. Verify Connection

You should see:
- Tools: add, multiply, greet
- Resources: config://server, status://health, file:///{path}
- Prompts: greeting, calculation

## Troubleshooting

### If you still see CORS errors:

#### 1. Check Browser Console

Open DevTools (F12) and look at the Console and Network tabs.

The error message will tell you exactly what's wrong:
- `Access-Control-Allow-Origin` mismatch
- Missing required headers
- Credential issues

#### 2. Verify Inspector Origin

Check what origin the Inspector is actually using:

```javascript
// In browser console
console.log(window.location.origin)
```

If it's not `http://localhost:6274`, start the server with that origin:
```bash
go run main.go -origin "http://YOUR_ACTUAL_ORIGIN"
```

#### 3. Check for Proxy/Middleware

MCP Inspector may use a proxy server. Check the Inspector docs for the actual request flow.

If Inspector uses `http://localhost:6277` as a proxy, that proxy needs to forward CORS headers correctly.

#### 4. Test Direct Connection

Test without Inspector to isolate the issue:

```javascript
// In browser console at http://localhost:6274
fetch('http://localhost:8080', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    jsonrpc: '2.0',
    method: 'tools/list',
    id: 1
  })
})
.then(r => r.json())
.then(console.log)
.catch(console.error)
```

If this works, the issue is with Inspector configuration, not the server.

#### 5. Enable Verbose Logging

Check the server logs when Inspector connects. You should see:
```
[MCP] POST / from http://localhost:6274
[MCP] Processing: method=initialize id=1
[MCP] Request processed: initialize → 200 OK
[MCP] POST / from http://localhost:6274
[MCP] Processing: method=notifications/initialized id=<nil>
[MCP] Notification processed: notifications/initialized → 202 Accepted
```

If you don't see these logs, Inspector isn't reaching your server.

### Common Issues

#### Issue: "CORS policy: No 'Access-Control-Allow-Origin' header"
**Solution**: Server not setting headers. Verify with curl that headers are present.

#### Issue: "Credential mode 'include' with wildcard origin"
**Solution**: This was fixed. Server now uses specific origin instead of `*`.

#### Issue: "preflight request doesn't pass access control check"
**Solution**: OPTIONS request failing. Check that OPTIONS returns 204 with proper headers.

#### Issue: Inspector shows "Connection refused"
**Solution**: Server not running or wrong port. Verify server is listening on :8080.

## Expected Network Flow

When Inspector connects, you should see this sequence:

1. **OPTIONS** (preflight for initialize)
   - Request: `OPTIONS http://localhost:8080`
   - Response: `204 No Content` with CORS headers

2. **POST** initialize
   - Request: `{"jsonrpc":"2.0","method":"initialize",...}`
   - Response: `200 OK` with CORS headers + result

3. **POST** notifications/initialized
   - Request: `{"jsonrpc":"2.0","method":"notifications/initialized"}`
   - Response: `202 Accepted` with CORS headers (no body)

4. **GET** (SSE stream)
   - Request: `GET http://localhost:8080` with `Accept: text/event-stream`
   - Response: `200 OK` with CORS headers + event stream

All responses should include these headers:
```
Access-Control-Allow-Origin: http://localhost:6274
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, Mcp-Session-Id, X-API-Key, Authorization, Last-Event-ID
Access-Control-Allow-Credentials: true
Access-Control-Expose-Headers: Mcp-Session-Id
```

## Still Not Working?

If CORS headers are present in curl but not in Inspector:

1. **Inspector may be using a proxy** - Check Inspector's network configuration
2. **Browser extension blocking** - Disable extensions and retry
3. **Inspector version** - Update to latest version (0.14.1+)
4. **Different Inspector origin** - Inspector may run on different port

Share the exact error message from browser console for more specific help.
