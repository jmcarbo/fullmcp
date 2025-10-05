# CORS Fix Summary

## Problem

MCP Inspector was getting CORS errors when sending `notifications/initialized` requests, even though `initialize` requests worked fine.

## Root Cause

**MCP Inspector sends `Referer` header instead of `Origin` header.**

The client request headers:
```
referer: http://localhost:6274/
```

But the server was only checking for:
```go
origin := r.Header.Get("Origin")  // Empty!
```

When `origin` was empty, the server fell back to wildcard `*`, but this creates an invalid combination when combined with `Access-Control-Allow-Credentials: true`.

## Solution

Added `Referer` header fallback support:

```go
origin := r.Header.Get("Origin")

// If no Origin header, try Referer as fallback
if origin == "" {
    if referer := r.Header.Get("Referer"); referer != "" {
        origin = extractOriginFromReferer(referer)  // Extract "http://localhost:6274"
    }
}
```

Now the server:
1. Checks for `Origin` header first
2. Falls back to `Referer` header if `Origin` is missing
3. Extracts the origin (scheme + host) from the Referer URL
4. Uses that specific origin instead of wildcard `*`
5. Sets `Access-Control-Allow-Credentials: true` (valid with specific origin)

## Changes Made

### 1. `transport/streamhttp/streamhttp.go`

#### Added `extractOriginFromReferer()` function (lines 589-618)
Extracts origin from referer URL:
- `http://localhost:6274/` → `http://localhost:6274`
- `https://example.com/path` → `https://example.com`

#### Updated `setCORSHeaders()` (lines 542-587)
- Added Referer fallback logic
- Added `Mcp-Protocol-Version` to allowed headers
- Added `Access-Control-Expose-Headers: Mcp-Session-Id`

#### Updated OPTIONS handler (lines 638-676)
- Same Referer fallback logic
- Same header additions

### 2. `transport/streamhttp/streamhttp_test.go`

#### Added tests (lines 489-539)
- `TestServer_CORSHeadersWithReferer` - Verifies Referer extraction works
- `TestExtractOriginFromReferer` - Tests origin extraction from various URLs

#### Updated existing test (line 98)
- Updated expected headers to include `Mcp-Protocol-Version`

### 3. `examples/streamhttp/main.go`

- Added debug logging to see Origin/Referer in requests
- Added README with testing instructions

## Test Results

### With Origin Header (Standard)
```bash
curl -H "Origin: http://localhost:6274" -X POST http://localhost:8080 \
  -d '{"jsonrpc":"2.0","method":"notifications/initialized"}'

✅ HTTP/1.1 202 Accepted
✅ Access-Control-Allow-Origin: http://localhost:6274
✅ Access-Control-Allow-Credentials: true
✅ Access-Control-Allow-Methods: GET, POST, OPTIONS
✅ Access-Control-Allow-Headers: Content-Type, Mcp-Session-Id, X-API-Key, Authorization, Last-Event-ID, Mcp-Protocol-Version
✅ Access-Control-Expose-Headers: Mcp-Session-Id
```

### With Referer Header (MCP Inspector)
```bash
curl -H "Referer: http://localhost:6274/" -X POST http://localhost:8080 \
  -d '{"jsonrpc":"2.0","method":"notifications/initialized"}'

✅ HTTP/1.1 202 Accepted
✅ Access-Control-Allow-Origin: http://localhost:6274
✅ Access-Control-Allow-Credentials: true
✅ Access-Control-Allow-Methods: GET, POST, OPTIONS
✅ Access-Control-Allow-Headers: Content-Type, Mcp-Session-Id, X-API-Key, Authorization, Last-Event-ID, Mcp-Protocol-Version
✅ Access-Control-Expose-Headers: Mcp-Session-Id
```

Both return **identical CORS headers** and work correctly!

## Testing

Run the server:
```bash
cd examples/streamhttp
go run main.go
```

Test with MCP Inspector headers:
```bash
./test-inspector-headers.sh
```

Connect with MCP Inspector:
1. Open http://localhost:6274
2. Connect to http://localhost:8080
3. All operations should work without CORS errors

## Compatibility

The fix maintains backward compatibility:
- ✅ Clients sending `Origin` header (standard) - works as before
- ✅ Clients sending `Referer` header (MCP Inspector) - now works
- ✅ Clients sending both - `Origin` takes precedence
- ✅ Clients sending neither - falls back to wildcard `*` (no credentials)

## Additional Headers

Also added support for:
- `Mcp-Protocol-Version` header (MCP 2025-06-18 spec)
- `Access-Control-Expose-Headers` for `Mcp-Session-Id`

This ensures clients can read the session ID from response headers.
