# Middleware

Middleware provides a composable way to add cross-cutting concerns to your MCP server, such as logging, authentication, rate limiting, and error recovery.

## Table of Contents

- [Overview](#overview)
- [Built-in Middleware](#built-in-middleware)
- [Creating Custom Middleware](#creating-custom-middleware)
- [Middleware Patterns](#middleware-patterns)
- [Best Practices](#best-practices)

## Overview

Middleware in FullMCP follows a chain-of-responsibility pattern where each middleware can:
- Pre-process incoming requests
- Call the next middleware in the chain
- Post-process responses
- Short-circuit the chain by returning early

### Middleware Signature

```go
type Middleware func(next Handler) Handler

type Handler func(ctx context.Context, req *Request) (*Response, error)
```

### Applying Middleware

```go
import "github.com/jmcarbo/fullmcp/server"

srv := server.New("my-server",
    server.WithMiddleware(
        server.RecoveryMiddleware(),
        LoggingMiddleware(),
        MetricsMiddleware(),
    ),
)
```

**Execution Order:**
Middleware executes in the order specified:
1. RecoveryMiddleware (outermost)
2. LoggingMiddleware
3. MetricsMiddleware
4. Actual handler (innermost)

## Built-in Middleware

### Recovery Middleware

Recovers from panics and returns errors gracefully:

```go
srv := server.New("my-server",
    server.WithMiddleware(
        server.RecoveryMiddleware(),
    ),
)
```

**Features:**
- Catches panics
- Logs stack trace
- Returns internal error to client
- Prevents server crashes

### Logging Middleware

Built-in logging middleware example:

```go
func LoggingMiddleware() server.Middleware {
    return func(next server.Handler) server.Handler {
        return func(ctx context.Context, req *server.Request) (*server.Response, error) {
            start := time.Now()

            log.Printf("[%s] %s - started", req.ID, req.Method)

            resp, err := next(ctx, req)

            duration := time.Since(start)
            status := "success"
            if err != nil {
                status = "error"
            }

            log.Printf("[%s] %s - %s (%v)", req.ID, req.Method, status, duration)

            return resp, err
        }
    }
}
```

## Creating Custom Middleware

### Basic Middleware Template

```go
func MyMiddleware() server.Middleware {
    return func(next server.Handler) server.Handler {
        return func(ctx context.Context, req *server.Request) (*server.Response, error) {
            // Pre-processing
            // ...

            // Call next middleware/handler
            resp, err := next(ctx, req)

            // Post-processing
            // ...

            return resp, err
        }
    }
}
```

### Authentication Middleware

```go
func AuthenticationMiddleware(apiKey string) server.Middleware {
    return func(next server.Handler) server.Handler {
        return func(ctx context.Context, req *server.Request) (*server.Response, error) {
            // Extract auth token from context or request metadata
            token, ok := ctx.Value("auth-token").(string)
            if !ok || token != apiKey {
                return nil, &mcp.Error{
                    Code:    mcp.ErrorCodeUnauthorized,
                    Message: "authentication required",
                }
            }

            return next(ctx, req)
        }
    }
}
```

### Rate Limiting Middleware

```go
import "golang.org/x/time/rate"

func RateLimitMiddleware(requestsPerSecond int) server.Middleware {
    limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), requestsPerSecond*2)

    return func(next server.Handler) server.Handler {
        return func(ctx context.Context, req *server.Request) (*server.Response, error) {
            if !limiter.Allow() {
                return nil, &mcp.Error{
                    Code:    mcp.ErrorCodeRateLimited,
                    Message: "rate limit exceeded",
                }
            }

            return next(ctx, req)
        }
    }
}
```

### Timeout Middleware

```go
func TimeoutMiddleware(timeout time.Duration) server.Middleware {
    return func(next server.Handler) server.Handler {
        return func(ctx context.Context, req *server.Request) (*server.Response, error) {
            ctx, cancel := context.WithTimeout(ctx, timeout)
            defer cancel()

            type result struct {
                resp *server.Response
                err  error
            }

            ch := make(chan result, 1)

            go func() {
                resp, err := next(ctx, req)
                ch <- result{resp, err}
            }()

            select {
            case res := <-ch:
                return res.resp, res.err
            case <-ctx.Done():
                return nil, &mcp.Error{
                    Code:    mcp.ErrorCodeTimeout,
                    Message: "request timeout",
                }
            }
        }
    }
}
```

### Metrics Middleware

```go
import "github.com/prometheus/client_golang/prometheus"

type MetricsMiddleware struct {
    requestsTotal   *prometheus.CounterVec
    requestDuration *prometheus.HistogramVec
}

func NewMetricsMiddleware() *MetricsMiddleware {
    mm := &MetricsMiddleware{
        requestsTotal: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "mcp_requests_total",
                Help: "Total number of MCP requests",
            },
            []string{"method", "status"},
        ),
        requestDuration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name:    "mcp_request_duration_seconds",
                Help:    "MCP request duration in seconds",
                Buckets: prometheus.DefBuckets,
            },
            []string{"method"},
        ),
    }

    prometheus.MustRegister(mm.requestsTotal)
    prometheus.MustRegister(mm.requestDuration)

    return mm
}

func (mm *MetricsMiddleware) Middleware() server.Middleware {
    return func(next server.Handler) server.Handler {
        return func(ctx context.Context, req *server.Request) (*server.Response, error) {
            start := time.Now()

            resp, err := next(ctx, req)

            duration := time.Since(start).Seconds()
            status := "success"
            if err != nil {
                status = "error"
            }

            mm.requestsTotal.WithLabelValues(req.Method, status).Inc()
            mm.requestDuration.WithLabelValues(req.Method).Observe(duration)

            return resp, err
        }
    }
}
```

### Request Validation Middleware

```go
func ValidationMiddleware() server.Middleware {
    return func(next server.Handler) server.Handler {
        return func(ctx context.Context, req *server.Request) (*server.Response, error) {
            // Validate request ID
            if req.ID == "" {
                return nil, &mcp.Error{
                    Code:    mcp.ErrorCodeInvalidRequest,
                    Message: "request ID is required",
                }
            }

            // Validate method
            validMethods := map[string]bool{
                "initialize":        true,
                "tools/list":        true,
                "tools/call":        true,
                "resources/list":    true,
                "resources/read":    true,
                "prompts/list":      true,
                "prompts/get":       true,
            }

            if !validMethods[req.Method] {
                return nil, &mcp.Error{
                    Code:    mcp.ErrorCodeMethodNotFound,
                    Message: fmt.Sprintf("unknown method: %s", req.Method),
                }
            }

            return next(ctx, req)
        }
    }
}
```

## Middleware Patterns

### Conditional Middleware

Execute middleware only for certain conditions:

```go
func ConditionalMiddleware(condition func(*server.Request) bool, middleware server.Middleware) server.Middleware {
    return func(next server.Handler) server.Handler {
        wrapped := middleware(next)

        return func(ctx context.Context, req *server.Request) (*server.Response, error) {
            if condition(req) {
                return wrapped(ctx, req)
            }
            return next(ctx, req)
        }
    }
}

// Usage
srv := server.New("my-server",
    server.WithMiddleware(
        ConditionalMiddleware(
            func(req *server.Request) bool {
                return strings.HasPrefix(req.Method, "tools/")
            },
            RateLimitMiddleware(10),
        ),
    ),
)
```

### Stateful Middleware

Middleware with internal state:

```go
type CacheMiddleware struct {
    cache map[string]*server.Response
    mu    sync.RWMutex
    ttl   time.Duration
}

func NewCacheMiddleware(ttl time.Duration) *CacheMiddleware {
    return &CacheMiddleware{
        cache: make(map[string]*server.Response),
        ttl:   ttl,
    }
}

func (cm *CacheMiddleware) Middleware() server.Middleware {
    return func(next server.Handler) server.Handler {
        return func(ctx context.Context, req *server.Request) (*server.Response, error) {
            // Only cache certain methods
            if req.Method != "resources/read" {
                return next(ctx, req)
            }

            // Generate cache key
            cacheKey := fmt.Sprintf("%s:%v", req.Method, req.Params)

            // Check cache
            cm.mu.RLock()
            if cached, ok := cm.cache[cacheKey]; ok {
                cm.mu.RUnlock()
                return cached, nil
            }
            cm.mu.RUnlock()

            // Execute request
            resp, err := next(ctx, req)
            if err != nil {
                return nil, err
            }

            // Store in cache
            cm.mu.Lock()
            cm.cache[cacheKey] = resp
            cm.mu.Unlock()

            // Set expiration
            time.AfterFunc(cm.ttl, func() {
                cm.mu.Lock()
                delete(cm.cache, cacheKey)
                cm.mu.Unlock()
            })

            return resp, nil
        }
    }
}
```

### Error Transformation Middleware

Transform errors before returning to client:

```go
func ErrorTransformMiddleware() server.Middleware {
    return func(next server.Handler) server.Handler {
        return func(ctx context.Context, req *server.Request) (*server.Response, error) {
            resp, err := next(ctx, req)
            if err == nil {
                return resp, nil
            }

            // Transform specific errors
            switch {
            case errors.Is(err, sql.ErrNoRows):
                return nil, &mcp.Error{
                    Code:    mcp.ErrorCodeResourceNotFound,
                    Message: "resource not found",
                }
            case errors.Is(err, context.DeadlineExceeded):
                return nil, &mcp.Error{
                    Code:    mcp.ErrorCodeTimeout,
                    Message: "request timeout",
                }
            default:
                // Pass through other errors
                return nil, err
            }
        }
    }
}
```

### Request Context Enrichment

Add data to context for downstream handlers:

```go
func ContextEnrichmentMiddleware() server.Middleware {
    return func(next server.Handler) server.Handler {
        return func(ctx context.Context, req *server.Request) (*server.Response, error) {
            // Add request ID to context
            ctx = context.WithValue(ctx, "request-id", req.ID)

            // Add timestamp
            ctx = context.WithValue(ctx, "start-time", time.Now())

            // Add trace ID
            traceID := generateTraceID()
            ctx = context.WithValue(ctx, "trace-id", traceID)

            return next(ctx, req)
        }
    }
}
```

### Retry Middleware

Automatically retry failed requests:

```go
func RetryMiddleware(maxRetries int, backoff time.Duration) server.Middleware {
    return func(next server.Handler) server.Handler {
        return func(ctx context.Context, req *server.Request) (*server.Response, error) {
            var resp *server.Response
            var err error

            for attempt := 0; attempt <= maxRetries; attempt++ {
                resp, err = next(ctx, req)

                // Success
                if err == nil {
                    return resp, nil
                }

                // Don't retry certain errors
                if mcpErr, ok := err.(*mcp.Error); ok {
                    if mcpErr.Code == mcp.ErrorCodeInvalidRequest {
                        return nil, err
                    }
                }

                // Last attempt
                if attempt == maxRetries {
                    break
                }

                // Wait before retry
                select {
                case <-ctx.Done():
                    return nil, ctx.Err()
                case <-time.After(backoff * time.Duration(attempt+1)):
                }
            }

            return resp, err
        }
    }
}
```

## Best Practices

### Middleware Order Matters

Place middleware in correct order:

```go
// ✅ Good: Recovery first, then logging, then auth
srv := server.New("my-server",
    server.WithMiddleware(
        server.RecoveryMiddleware(),  // Catch all panics
        LoggingMiddleware(),           // Log all requests
        AuthenticationMiddleware(),    // Authenticate
        RateLimitMiddleware(100),      // Rate limit
        ValidationMiddleware(),        // Validate
    ),
)

// ❌ Bad: Auth before recovery
srv := server.New("my-server",
    server.WithMiddleware(
        AuthenticationMiddleware(),   // Will panic before recovery
        server.RecoveryMiddleware(),
    ),
)
```

### Keep Middleware Focused

Each middleware should do one thing:

```go
// ✅ Good: Separate concerns
LoggingMiddleware()
AuthenticationMiddleware()
MetricsMiddleware()

// ❌ Bad: Everything in one middleware
MonolithicMiddleware()  // Does logging, auth, metrics, etc.
```

### Use Context for Passing Data

```go
// ✅ Good: Use context
func AuthMiddleware() server.Middleware {
    return func(next server.Handler) server.Handler {
        return func(ctx context.Context, req *server.Request) (*server.Response, error) {
            user := authenticateUser(req)
            ctx = context.WithValue(ctx, "user", user)
            return next(ctx, req)
        }
    }
}

// ❌ Bad: Modify request
func AuthMiddleware() server.Middleware {
    return func(next server.Handler) server.Handler {
        return func(ctx context.Context, req *server.Request) (*server.Response, error) {
            req.User = authenticateUser(req)  // Don't modify request
            return next(ctx, req)
        }
    }
}
```

### Handle Errors Gracefully

```go
// ✅ Good: Proper error handling
func MyMiddleware() server.Middleware {
    return func(next server.Handler) server.Handler {
        return func(ctx context.Context, req *server.Request) (*server.Response, error) {
            if err := validate(req); err != nil {
                return nil, &mcp.Error{
                    Code:    mcp.ErrorCodeInvalidRequest,
                    Message: err.Error(),
                }
            }

            return next(ctx, req)
        }
    }
}
```

### Resource Cleanup

```go
func ResourceMiddleware() server.Middleware {
    return func(next server.Handler) server.Handler {
        return func(ctx context.Context, req *server.Request) (*server.Response, error) {
            resource := acquireResource()
            defer releaseResource(resource)

            ctx = context.WithValue(ctx, "resource", resource)
            return next(ctx, req)
        }
    }
}
```

### Testing Middleware

```go
func TestLoggingMiddleware(t *testing.T) {
    // Create test handler
    handler := func(ctx context.Context, req *server.Request) (*server.Response, error) {
        return &server.Response{Result: "ok"}, nil
    }

    // Wrap with middleware
    middleware := LoggingMiddleware()
    wrapped := middleware(handler)

    // Test
    req := &server.Request{
        ID:     "test-1",
        Method: "tools/list",
    }

    resp, err := wrapped(context.Background(), req)

    assert.NoError(t, err)
    assert.Equal(t, "ok", resp.Result)
}
```

## Example: Complete Middleware Stack

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/jmcarbo/fullmcp/server"
)

func main() {
    // Create metrics middleware
    metrics := NewMetricsMiddleware()

    // Create cache middleware
    cache := NewCacheMiddleware(5 * time.Minute)

    // Create server with middleware stack
    srv := server.New("production-server",
        server.WithMiddleware(
            // 1. Recover from panics (outermost)
            server.RecoveryMiddleware(),

            // 2. Log all requests
            LoggingMiddleware(),

            // 3. Collect metrics
            metrics.Middleware(),

            // 4. Authenticate requests
            AuthenticationMiddleware(),

            // 5. Rate limit
            RateLimitMiddleware(100),

            // 6. Add request timeouts
            TimeoutMiddleware(30 * time.Second),

            // 7. Validate requests
            ValidationMiddleware(),

            // 8. Cache responses
            cache.Middleware(),

            // 9. Transform errors
            ErrorTransformMiddleware(),

            // 10. Retry on failure
            RetryMiddleware(3, time.Second),
        ),
    )

    // Add tools, resources, prompts...

    log.Fatal(srv.Run(context.Background()))
}
```

## Related Documentation

- [Architecture Overview](./architecture.md)
- [Authentication](./authentication.md)
- [Transports](./transports.md)
