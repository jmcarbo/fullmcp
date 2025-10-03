# Performance Benchmarks

FullMCP is designed for high performance with minimal overhead. This document provides benchmark results and optimization guidelines.

## Table of Contents

- [Benchmark Results](#benchmark-results)
- [Running Benchmarks](#running-benchmarks)
- [Optimization Techniques](#optimization-techniques)
- [Performance Profiling](#performance-profiling)
- [Best Practices](#best-practices)

## Benchmark Results

Performance benchmarks on Apple M-series processor (results may vary by platform):

### Core Operations

| Operation | Time | Memory | Allocations |
|-----------|------|--------|-------------|
| Tool call (simple) | ~33 ns/op | 40 B/op | 1 allocs/op |
| Tool call (with reflection) | ~450 ns/op | 240 B/op | 6 allocs/op |
| Resource read (static) | ~22 ns/op | 16 B/op | 1 allocs/op |
| Resource read (template) | ~180 ns/op | 128 B/op | 4 allocs/op |
| Prompt render (simple) | ~95 ns/op | 64 B/op | 2 allocs/op |
| Prompt render (complex) | ~520 ns/op | 384 B/op | 8 allocs/op |

### Message Handling

| Operation | Time | Memory | Allocations |
|-----------|------|--------|-------------|
| JSON-RPC encode | ~850 ns/op | 512 B/op | 6 allocs/op |
| JSON-RPC decode | ~920 ns/op | 576 B/op | 12 allocs/op |
| Message routing | ~65 ns/op | 32 B/op | 1 allocs/op |
| Full request cycle | ~1.3 μs/op | 1.2 KB/op | 18 allocs/op |

### Client Operations

| Operation | Time | Memory | Allocations |
|-----------|------|--------|-------------|
| Client connection setup | ~25 μs/op | 18 KB/op | 145 allocs/op |
| Request/response roundtrip | ~1.8 μs/op | 1.5 KB/op | 22 allocs/op |
| Concurrent requests (10) | ~12 μs/op | 15 KB/op | 220 allocs/op |

### Transport Performance

| Transport | Latency | Throughput | Memory |
|-----------|---------|------------|--------|
| stdio | ~50 μs | 20K req/s | 2 KB/req |
| HTTP (local) | ~200 μs | 5K req/s | 4 KB/req |
| WebSocket | ~120 μs | 8K req/s | 3 KB/req |

## Running Benchmarks

### All Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem ./...

# Run with more iterations for accuracy
go test -bench=. -benchmem -benchtime=10s ./...

# Save results to file
go test -bench=. -benchmem ./... | tee bench.txt
```

### Specific Package Benchmarks

```bash
# Server benchmarks
go test -bench=. -benchmem ./server

# Client benchmarks
go test -bench=. -benchmem ./client

# Builder benchmarks
go test -bench=. -benchmem ./builder

# Transport benchmarks
go test -bench=. -benchmem ./transport/...
```

### Specific Benchmark

```bash
# Run specific benchmark by name
go test -bench=BenchmarkToolCall -benchmem ./server

# Run benchmarks matching pattern
go test -bench=Tool -benchmem ./...
```

### Compare Benchmarks

Use `benchstat` to compare results:

```bash
# Install benchstat
go install golang.org/x/perf/cmd/benchstat@latest

# Run and save baseline
go test -bench=. -benchmem ./... > old.txt

# Make changes...

# Run and save new results
go test -bench=. -benchmem ./... > new.txt

# Compare
benchstat old.txt new.txt
```

Example output:
```
name           old time/op    new time/op    delta
ToolCall-8       450ns ± 2%      33ns ± 1%   -92.67%  (p=0.000 n=10+10)

name           old alloc/op   new alloc/op   delta
ToolCall-8        240B ± 0%       40B ± 0%   -83.33%  (p=0.000 n=10+10)

name           old allocs/op  new allocs/op  delta
ToolCall-8        6.00 ± 0%      1.00 ± 0%   -83.33%  (p=0.000 n=10+10)
```

## Optimization Techniques

### 1. Reduce Allocations

**Before:**
```go
func (tm *ToolManager) CallTool(name string, args []byte) (interface{}, error) {
    // Creates new map on each call
    tools := make(map[string]*Tool)
    for _, t := range tm.tools {
        tools[t.Name] = t
    }
    // ...
}
```

**After:**
```go
func (tm *ToolManager) CallTool(name string, args []byte) (interface{}, error) {
    // Use existing map with read lock
    tm.mu.RLock()
    tool, exists := tm.tools[name]
    tm.mu.RUnlock()
    // ...
}
```

### 2. Pool Frequently Allocated Objects

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func processMessage(msg []byte) ([]byte, error) {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()

    // Use buffer...
    return buf.Bytes(), nil
}
```

### 3. Use String Interning

```go
type stringInterner struct {
    strings sync.Map
}

func (si *stringInterner) Intern(s string) string {
    if v, ok := si.strings.Load(s); ok {
        return v.(string)
    }

    si.strings.Store(s, s)
    return s
}

// Use for frequently repeated strings
var interner = &stringInterner{}

func parseRequest(data []byte) *Request {
    req := &Request{
        Method: interner.Intern(extractMethod(data)),
    }
    return req
}
```

### 4. Optimize JSON Encoding

```go
// Use json.Encoder for streaming
func writeResponse(w io.Writer, resp *Response) error {
    enc := json.NewEncoder(w)
    return enc.Encode(resp)
}

// Pre-allocate buffer for known sizes
func marshalMessage(msg *Message) ([]byte, error) {
    buf := make([]byte, 0, 1024) // Estimated size
    data, err := json.Marshal(msg)
    buf = append(buf, data...)
    return buf, err
}
```

### 5. Cache Computed Values

```go
type CachedTool struct {
    *Tool
    schemaCache []byte
    schemaMu    sync.RWMutex
}

func (ct *CachedTool) GetSchema() ([]byte, error) {
    ct.schemaMu.RLock()
    if ct.schemaCache != nil {
        defer ct.schemaMu.RUnlock()
        return ct.schemaCache, nil
    }
    ct.schemaMu.RUnlock()

    ct.schemaMu.Lock()
    defer ct.schemaMu.Unlock()

    // Double-check after acquiring write lock
    if ct.schemaCache != nil {
        return ct.schemaCache, nil
    }

    schema, err := json.Marshal(ct.InputSchema)
    if err != nil {
        return nil, err
    }

    ct.schemaCache = schema
    return schema, nil
}
```

## Performance Profiling

### CPU Profiling

```bash
# Run with CPU profiling
go test -bench=BenchmarkToolCall -cpuprofile=cpu.prof ./server

# Analyze profile
go tool pprof cpu.prof

# Common pprof commands:
# top10 - Show top 10 functions by CPU time
# list <function> - Show source code with annotations
# web - Open visualization in browser
```

### Memory Profiling

```bash
# Run with memory profiling
go test -bench=BenchmarkToolCall -memprofile=mem.prof ./server

# Analyze profile
go tool pprof mem.prof

# Show allocations
go tool pprof -alloc_space mem.prof
```

### Trace Analysis

```bash
# Generate trace
go test -bench=BenchmarkToolCall -trace=trace.out ./server

# View trace
go tool trace trace.out
```

### Continuous Profiling

```go
import (
    "net/http"
    _ "net/http/pprof"
)

func main() {
    // Start pprof server
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

    // Run your server...
}

// Access profiles:
// http://localhost:6060/debug/pprof/
// http://localhost:6060/debug/pprof/heap
// http://localhost:6060/debug/pprof/goroutine
```

## Best Practices

### 1. Benchmark What Matters

Focus on hot paths:

```go
// ✅ Good: Benchmark critical path
func BenchmarkToolCall(b *testing.B) {
    srv := setupServer()
    args := []byte(`{"a": 5, "b": 3}`)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        srv.CallTool(context.Background(), "add", args)
    }
}

// ❌ Bad: Benchmark includes setup
func BenchmarkToolCall(b *testing.B) {
    for i := 0; i < b.N; i++ {
        srv := setupServer() // Setup shouldn't be measured
        srv.CallTool(context.Background(), "add", []byte(`{"a": 5, "b": 3}`))
    }
}
```

### 2. Use b.ResetTimer()

Exclude setup from measurements:

```go
func BenchmarkComplexOperation(b *testing.B) {
    // Setup (not measured)
    srv := setupServer()
    data := loadTestData()

    b.ResetTimer() // Start timing here

    for i := 0; i < b.N; i++ {
        srv.Process(data)
    }
}
```

### 3. Avoid Compiler Optimizations

Prevent dead code elimination:

```go
var result interface{}

func BenchmarkOperation(b *testing.B) {
    var r interface{}

    for i := 0; i < b.N; i++ {
        r = doOperation()
    }

    result = r // Prevent optimization
}
```

### 4. Run Multiple Times

Get statistically significant results:

```go
# Run 10 times to get stable results
go test -bench=. -count=10 ./... | tee bench.txt
benchstat bench.txt
```

### 5. Parallel Benchmarks

Test concurrent performance:

```go
func BenchmarkToolCallParallel(b *testing.B) {
    srv := setupServer()
    args := []byte(`{"a": 5, "b": 3}`)

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            srv.CallTool(context.Background(), "add", args)
        }
    })
}
```

### 6. Sub-Benchmarks

Compare different implementations:

```go
func BenchmarkJSONEncodings(b *testing.B) {
    data := getTestData()

    b.Run("Marshal", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            json.Marshal(data)
        }
    })

    b.Run("Encoder", func(b *testing.B) {
        var buf bytes.Buffer
        enc := json.NewEncoder(&buf)
        b.ResetTimer()

        for i := 0; i < b.N; i++ {
            buf.Reset()
            enc.Encode(data)
        }
    })
}
```

## Performance Targets

### Latency Targets

| Operation | Target | Current |
|-----------|--------|---------|
| Tool call | < 100 ns | 33 ns ✅ |
| Resource read | < 50 ns | 22 ns ✅ |
| Message handling | < 2 μs | 1.3 μs ✅ |
| Client roundtrip | < 5 μs | 1.8 μs ✅ |

### Throughput Targets

| Component | Target | Current |
|-----------|--------|---------|
| Tool calls | > 1M/s | 30M/s ✅ |
| Message handling | > 500K/s | 770K/s ✅ |
| stdio transport | > 10K/s | 20K/s ✅ |

### Memory Targets

| Operation | Target | Current |
|-----------|--------|---------|
| Tool call | < 100 B | 40 B ✅ |
| Message handling | < 2 KB | 1.2 KB ✅ |
| Client connection | < 25 KB | 18 KB ✅ |

## Regression Detection

Set up automated benchmark regression detection:

```bash
#!/bin/bash
# bench-regression.sh

# Run baseline
git checkout main
go test -bench=. -benchmem ./... > baseline.txt

# Run current
git checkout feature-branch
go test -bench=. -benchmem ./... > current.txt

# Compare
benchstat baseline.txt current.txt > comparison.txt

# Check for regressions > 10%
if grep -q "~" comparison.txt; then
    echo "Performance regression detected!"
    cat comparison.txt
    exit 1
fi
```

## Optimization Checklist

Before claiming optimization:

- [ ] Run benchmarks with `-benchtime=10s`
- [ ] Run with `-count=10` for statistical significance
- [ ] Use `benchstat` to verify improvements
- [ ] Profile with `pprof` to identify bottlenecks
- [ ] Check for reduced allocations with `-benchmem`
- [ ] Test under concurrent load with `RunParallel`
- [ ] Verify no correctness regressions with `go test`
- [ ] Document optimization rationale

## Related Documentation

- [Architecture Overview](./architecture.md)
- [Contributing Guidelines](../CONTRIBUTING.md)
