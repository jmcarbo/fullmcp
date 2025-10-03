# Managing Resources

Resources in MCP provide read-only access to data sources. FullMCP supports both static resources and dynamic resource templates with URI-based parameter extraction.

## Table of Contents

- [Overview](#overview)
- [Static Resources](#static-resources)
- [Resource Templates](#resource-templates)
- [Resource Metadata](#resource-metadata)
- [Content Types](#content-types)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

Resources are identified by URIs and can contain any type of data:
- Configuration files
- Documentation
- Database queries
- File system access
- API responses
- Computed data

**Key Characteristics:**
- Read-only by design
- URI-addressable
- Support MIME types
- Can be static or templated
- Include metadata for versioning and targeting

## Static Resources

Static resources provide fixed data accessible via a URI.

### Basic Static Resource

```go
import (
    "context"
    "encoding/json"
    "github.com/jmcarbo/fullmcp/builder"
    "github.com/jmcarbo/fullmcp/server"
)

func main() {
    srv := server.New("config-server")

    // Create static resource
    configResource := builder.NewResource("config://app").
        Name("Application Config").
        Description("Main application configuration").
        MimeType("application/json").
        Reader(func(ctx context.Context) ([]byte, error) {
            config := map[string]interface{}{
                "debug":      true,
                "port":       8080,
                "database":   "postgres://localhost/mydb",
                "maxRetries": 3,
            }
            return json.Marshal(config)
        }).
        Build()

    srv.AddResource(configResource)
}
```

### Builder Methods

#### Name (required)

```go
func (rb *ResourceBuilder) Name(name string) *ResourceBuilder
```

Human-readable resource name.

#### URI (set via constructor)

```go
func NewResource(uri string) *ResourceBuilder
```

Unique resource identifier. Common schemes:
- `config://`: Configuration data
- `file://`: File system access
- `db://`: Database queries
- `api://`: API endpoints
- Custom schemes for your domain

#### Title

```go
func (rb *ResourceBuilder) Title(title string) *ResourceBuilder
```

Human-friendly display name (MCP 2025-06-18):

```go
resource := builder.NewResource("config://database").
    Name("database-config").
    Title("Database Configuration").
    // ...
```

#### Description

```go
func (rb *ResourceBuilder) Description(desc string) *ResourceBuilder
```

Detailed resource description.

#### MimeType

```go
func (rb *ResourceBuilder) MimeType(mimeType string) *ResourceBuilder
```

Content MIME type:
- `application/json`
- `text/plain`
- `text/html`
- `text/markdown`
- `application/xml`
- `image/png`
- Custom types

#### Reader

```go
func (rb *ResourceBuilder) Reader(fn func(context.Context) ([]byte, error)) *ResourceBuilder
```

Function that returns resource content as bytes.

### Complex Static Resources

#### Database Query Resource

```go
type DatabaseResource struct {
    db *sql.DB
}

func (dr *DatabaseResource) CreateResource() *mcp.Resource {
    return builder.NewResource("db://users/count").
        Name("User Count").
        Description("Total number of users in the database").
        MimeType("application/json").
        Reader(func(ctx context.Context) ([]byte, error) {
            var count int
            err := dr.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
            if err != nil {
                return nil, err
            }

            result := map[string]interface{}{
                "count":     count,
                "timestamp": time.Now().Unix(),
            }
            return json.Marshal(result)
        }).
        Build()
}
```

#### API Response Resource

```go
resource := builder.NewResource("api://weather/current").
    Name("Current Weather").
    Description("Latest weather data from external API").
    MimeType("application/json").
    Reader(func(ctx context.Context) ([]byte, error) {
        resp, err := http.Get("https://api.weather.com/current")
        if err != nil {
            return nil, err
        }
        defer resp.Body.Close()

        return ioutil.ReadAll(resp.Body)
    }).
    Build()
```

## Resource Templates

Resource templates allow parameterized resources using URI templates.

### Basic Template

```go
// Template: file:///{path}
// Matches: file:///config.json, file:///data/users.json

fileTemplate := builder.NewResourceTemplate("file:///{path}").
    Name("File Reader").
    Description("Read files from the filesystem").
    MimeType("text/plain").
    ReaderSimple(func(ctx context.Context, path string) ([]byte, error) {
        return os.ReadFile(path)
    }).
    Build()

srv.AddResourceTemplate(fileTemplate)
```

### URI Template Syntax

URI templates use `{parameter}` syntax:

```go
// Single parameter
"user:///{id}"                  // Matches: user:///123

// Multiple parameters
"posts:///{author}/{postId}"    // Matches: posts:///john/456

// Nested paths
"files:///{dir}/{filename}"     // Matches: files:///config/app.json
```

### Template Reader Functions

#### ReaderSimple (for single parameter)

```go
func (rtb *ResourceTemplateBuilder) ReaderSimple(
    fn func(context.Context, string) ([]byte, error)
) *ResourceTemplateBuilder
```

For templates with one parameter:

```go
template := builder.NewResourceTemplate("doc:///{page}").
    ReaderSimple(func(ctx context.Context, page string) ([]byte, error) {
        return loadDocumentation(page)
    }).
    Build()
```

#### Reader (for multiple parameters)

```go
func (rtb *ResourceTemplateBuilder) Reader(
    fn func(context.Context, map[string]string) ([]byte, error)
) *ResourceTemplateBuilder
```

For templates with multiple parameters:

```go
template := builder.NewResourceTemplate("data:///{table}/{id}").
    Reader(func(ctx context.Context, params map[string]string) ([]byte, error) {
        table := params["table"]
        id := params["id"]

        var data []byte
        err := db.QueryRowContext(ctx,
            fmt.Sprintf("SELECT data FROM %s WHERE id = ?", table),
            id,
        ).Scan(&data)

        return data, err
    }).
    Build()
```

### Advanced Template Examples

#### Multi-parameter Template

```go
// logs:///{service}/{date}/{level}
logsTemplate := builder.NewResourceTemplate("logs:///{service}/{date}/{level}").
    Name("Service Logs").
    Description("Access service logs by date and level").
    MimeType("text/plain").
    Reader(func(ctx context.Context, params map[string]string) ([]byte, error) {
        service := params["service"]
        date := params["date"]
        level := params["level"]

        logFile := fmt.Sprintf("/var/log/%s/%s/%s.log", service, date, level)
        return os.ReadFile(logFile)
    }).
    Build()

// Access: logs:///api/2025-01-15/error
```

#### Database Table Template

```go
tableTemplate := builder.NewResourceTemplate("db:///{table}").
    Name("Database Table").
    Description("Query all rows from a database table").
    MimeType("application/json").
    ReaderSimple(func(ctx context.Context, table string) ([]byte, error) {
        // Validate table name to prevent SQL injection
        validTables := map[string]bool{
            "users": true, "posts": true, "comments": true,
        }
        if !validTables[table] {
            return nil, fmt.Errorf("invalid table: %s", table)
        }

        rows, err := db.QueryContext(ctx,
            fmt.Sprintf("SELECT * FROM %s LIMIT 100", table),
        )
        if err != nil {
            return nil, err
        }
        defer rows.Close()

        // Convert rows to JSON
        var results []map[string]interface{}
        cols, _ := rows.Columns()

        for rows.Next() {
            columns := make([]interface{}, len(cols))
            columnPointers := make([]interface{}, len(cols))
            for i := range columns {
                columnPointers[i] = &columns[i]
            }

            rows.Scan(columnPointers...)

            row := make(map[string]interface{})
            for i, col := range cols {
                row[col] = columns[i]
            }
            results = append(results, row)
        }

        return json.Marshal(results)
    }).
    Build()
```

## Resource Metadata

Metadata provides version tracking and audience targeting (MCP 2025-06-18).

### Version Tracking

```go
resource := builder.NewResource("config://app").
    Name("App Config").
    MimeType("application/json").
    Meta(map[string]interface{}{
        "version": "1.2.3",
        "updated": "2025-01-15T10:30:00Z",
        "author":  "config-service",
    }).
    Reader(func(ctx context.Context) ([]byte, error) {
        // Return config
    }).
    Build()
```

### Audience Targeting

```go
resource := builder.NewResource("docs://internal/api").
    Name("Internal API Docs").
    Meta(map[string]interface{}{
        "audience": "internal",
        "security": "confidential",
    }).
    Reader(func(ctx context.Context) ([]byte, error) {
        // Return docs
    }).
    Build()
```

### Common Metadata Fields

```go
Meta(map[string]interface{}{
    "version":      "1.0.0",         // Resource version
    "updated":      "2025-01-15",    // Last update timestamp
    "author":       "system",         // Creator/owner
    "audience":     "public",         // public, internal, admin
    "tags":         []string{"api", "config"},
    "deprecated":   false,
    "replacedBy":   "config://app/v2",
    "cacheControl": "max-age=3600",
})
```

## Content Types

### JSON Resources

```go
resource := builder.NewResource("data://stats").
    MimeType("application/json").
    Reader(func(ctx context.Context) ([]byte, error) {
        stats := map[string]interface{}{
            "users":    1234,
            "requests": 56789,
            "uptime":   "99.9%",
        }
        return json.Marshal(stats)
    }).
    Build()
```

### Plain Text Resources

```go
resource := builder.NewResource("docs://readme").
    MimeType("text/plain").
    Reader(func(ctx context.Context) ([]byte, error) {
        return []byte("# README\n\nWelcome to the application!"), nil
    }).
    Build()
```

### Markdown Resources

```go
resource := builder.NewResource("docs://guide").
    MimeType("text/markdown").
    Reader(func(ctx context.Context) ([]byte, error) {
        return os.ReadFile("docs/guide.md")
    }).
    Build()
```

### Binary Resources

```go
resource := builder.NewResource("images://logo").
    MimeType("image/png").
    Reader(func(ctx context.Context) ([]byte, error) {
        return os.ReadFile("assets/logo.png")
    }).
    Build()
```

## Best Practices

### URI Naming Conventions

✅ **Good:**
```go
"config://app"
"db://users/count"
"file:///config/app.json"
"api://weather/current"
```

❌ **Bad:**
```go
"my-config"              // Missing scheme
"config://app config"    // Spaces
"CONFIG://APP"           // Uppercase (use lowercase)
```

### Error Handling

Always handle errors gracefully:

```go
Reader(func(ctx context.Context) ([]byte, error) {
    data, err := fetchData(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch data: %w", err)
    }

    if len(data) == 0 {
        return nil, &mcp.Error{
            Code:    mcp.ErrorCodeResourceNotFound,
            Message: "no data available",
        }
    }

    return data, nil
})
```

### Context Usage

Respect context cancellation:

```go
Reader(func(ctx context.Context) ([]byte, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        return fetchData()
    }
})
```

### Security

#### Validate Template Parameters

```go
ReaderSimple(func(ctx context.Context, path string) ([]byte, error) {
    // Prevent path traversal
    if strings.Contains(path, "..") {
        return nil, fmt.Errorf("invalid path")
    }

    // Whitelist approach
    allowedPaths := []string{"/config", "/data", "/docs"}
    valid := false
    for _, allowed := range allowedPaths {
        if strings.HasPrefix(path, allowed) {
            valid = true
            break
        }
    }

    if !valid {
        return nil, fmt.Errorf("path not allowed")
    }

    return os.ReadFile(path)
})
```

#### Sanitize Database Queries

```go
ReaderSimple(func(ctx context.Context, table string) ([]byte, error) {
    // Use parameterized queries
    // Validate table names against whitelist
    validTables := map[string]bool{
        "users": true,
        "posts": true,
    }

    if !validTables[table] {
        return nil, fmt.Errorf("invalid table")
    }

    // Safe to use table name
    query := fmt.Sprintf("SELECT * FROM %s", table)
    // ...
})
```

### Performance

#### Caching

```go
type CachedResource struct {
    cache     map[string][]byte
    cacheMux  sync.RWMutex
    ttl       time.Duration
    timestamps map[string]time.Time
}

func (cr *CachedResource) Reader(ctx context.Context) ([]byte, error) {
    key := "resource-key"

    // Check cache
    cr.cacheMux.RLock()
    if data, ok := cr.cache[key]; ok {
        if time.Since(cr.timestamps[key]) < cr.ttl {
            cr.cacheMux.RUnlock()
            return data, nil
        }
    }
    cr.cacheMux.RUnlock()

    // Fetch fresh data
    data, err := cr.fetchData(ctx)
    if err != nil {
        return nil, err
    }

    // Update cache
    cr.cacheMux.Lock()
    cr.cache[key] = data
    cr.timestamps[key] = time.Now()
    cr.cacheMux.Unlock()

    return data, nil
}
```

#### Streaming Large Resources

```go
// For very large resources, consider streaming
// Note: Current MCP spec returns full bytes, but plan accordingly
Reader(func(ctx context.Context) ([]byte, error) {
    var buf bytes.Buffer
    writer := gzip.NewWriter(&buf)

    // Write compressed data
    _, err := writer.Write(largeData)
    writer.Close()

    return buf.Bytes(), err
})
```

## Testing Resources

### Unit Testing

```go
func TestResource(t *testing.T) {
    resource := builder.NewResource("test://data").
        Name("Test Data").
        MimeType("application/json").
        Reader(func(ctx context.Context) ([]byte, error) {
            return json.Marshal(map[string]string{"key": "value"})
        }).
        Build()

    require.NotNil(t, resource)
    assert.Equal(t, "test://data", resource.URI)
    assert.Equal(t, "application/json", resource.MimeType)
}
```

### Integration Testing

```go
func TestResourceRead(t *testing.T) {
    srv := server.New("test")

    resource := builder.NewResource("config://test").
        Name("Test Config").
        MimeType("application/json").
        Reader(func(ctx context.Context) ([]byte, error) {
            return []byte(`{"test": true}`), nil
        }).
        Build()

    srv.AddResource(resource)

    // Read resource
    content, err := srv.ResourceManager().ReadResource(
        context.Background(),
        "config://test",
    )

    require.NoError(t, err)
    assert.Equal(t, "application/json", content.MimeType)
    assert.JSONEq(t, `{"test": true}`, content.Text)
}
```

## Examples

### Configuration Management

```go
// Base configuration
srv.AddResource(builder.NewResource("config://base").
    Name("Base Config").
    MimeType("application/json").
    Reader(func(ctx context.Context) ([]byte, error) {
        return json.Marshal(baseConfig)
    }).
    Build())

// Environment-specific configs
for _, env := range []string{"dev", "staging", "prod"} {
    cfg := loadConfig(env)
    srv.AddResource(builder.NewResource(fmt.Sprintf("config://%s", env)).
        Name(fmt.Sprintf("%s Config", env)).
        MimeType("application/json").
        Reader(func(ctx context.Context) ([]byte, error) {
            return json.Marshal(cfg)
        }).
        Build())
}
```

### Documentation Server

```go
// Template for markdown docs
srv.AddResourceTemplate(builder.NewResourceTemplate("docs:///{path}").
    Name("Documentation").
    MimeType("text/markdown").
    ReaderSimple(func(ctx context.Context, path string) ([]byte, error) {
        return os.ReadFile(filepath.Join("docs", path+".md"))
    }).
    Build())
```

See [`examples/advanced-server`](../examples/advanced-server/) for complete examples.

## Related Documentation

- [Architecture Overview](./architecture.md)
- [Building Tools](./tools.md)
- [Server API Reference](https://pkg.go.dev/github.com/jmcarbo/fullmcp/server)
