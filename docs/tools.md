# Building Tools

Tools are the primary way to expose functionality through the Model Context Protocol. FullMCP provides a type-safe, reflection-based approach to building tools with automatic JSON schema generation.

## Table of Contents

- [Quick Start](#quick-start)
- [Tool Builder API](#tool-builder-api)
- [Input Schemas](#input-schemas)
- [Output Schemas](#output-schemas)
- [Tool Hints](#tool-hints)
- [Error Handling](#error-handling)
- [Advanced Patterns](#advanced-patterns)
- [Best Practices](#best-practices)

## Quick Start

### Basic Tool

```go
package main

import (
    "context"
    "github.com/jmcarbo/fullmcp/builder"
    "github.com/jmcarbo/fullmcp/server"
)

type AddArgs struct {
    A int `json:"a" jsonschema:"required,description=First number"`
    B int `json:"b" jsonschema:"required,description=Second number"`
}

func main() {
    srv := server.New("calculator")

    addTool, _ := builder.NewTool("add").
        Description("Add two numbers").
        Handler(func(ctx context.Context, args AddArgs) (int, error) {
            return args.A + args.B, nil
        }).
        Build()

    srv.AddTool(addTool)
}
```

## Tool Builder API

### Creating a Tool

```go
func NewTool(name string) *ToolBuilder
```

Tool names must be unique within a server and follow these conventions:
- Use lowercase with underscores: `get_user`, `calculate_sum`
- Be descriptive and action-oriented
- Avoid generic names like `process` or `handle`

### Builder Methods

#### Description

```go
func (tb *ToolBuilder) Description(desc string) *ToolBuilder
```

Provide a clear description of what the tool does:

```go
tool, _ := builder.NewTool("search_files").
    Description("Search for files matching a pattern in a directory").
    // ...
```

#### Title

```go
func (tb *ToolBuilder) Title(title string) *ToolBuilder
```

Human-friendly display name (MCP 2025-06-18):

```go
tool, _ := builder.NewTool("get_weather").
    Title("Get Weather Forecast").
    Description("Retrieve current weather and forecast for a location").
    // ...
```

#### Handler

```go
func (tb *ToolBuilder) Handler(fn interface{}) *ToolBuilder
```

The handler function must follow this signature:
```go
func(ctx context.Context, args YourArgsType) (ResultType, error)
```

**Requirements:**
- First parameter must be `context.Context`
- Second parameter is your input struct (schema auto-generated)
- Return type can be any serializable type
- Must return error as second value

## Input Schemas

Input schemas are automatically generated from Go struct tags using the `jsonschema` package.

### Basic Types

```go
type UserArgs struct {
    Name    string  `json:"name" jsonschema:"required,description=User's full name"`
    Age     int     `json:"age" jsonschema:"description=User's age"`
    Email   string  `json:"email" jsonschema:"format=email"`
    Score   float64 `json:"score" jsonschema:"minimum=0,maximum=100"`
    Active  bool    `json:"active"`
}
```

### Struct Tags

#### `json` tag
- Required for field serialization
- Defines JSON property name

#### `jsonschema` tag

**Common validators:**
- `required`: Field is required
- `description=<text>`: Field description
- `minimum=<n>`: Minimum value (numbers)
- `maximum=<n>`: Maximum value (numbers)
- `minLength=<n>`: Minimum string length
- `maxLength=<n>`: Maximum string length
- `pattern=<regex>`: String pattern validation
- `format=<type>`: String format (email, uri, date-time, etc.)
- `enum=<val1>|<val2>`: Enumeration of allowed values

### Enumerations

```go
type OperationArgs struct {
    Operation string `json:"op" jsonschema:"required,enum=add|subtract|multiply|divide"`
    A         float64 `json:"a" jsonschema:"required"`
    B         float64 `json:"b" jsonschema:"required"`
}
```

### Arrays

```go
type BatchArgs struct {
    Items []string `json:"items" jsonschema:"required,minItems=1,maxItems=100"`
}
```

### Nested Objects

```go
type Address struct {
    Street string `json:"street" jsonschema:"required"`
    City   string `json:"city" jsonschema:"required"`
    Zip    string `json:"zip" jsonschema:"pattern=^[0-9]{5}$"`
}

type CreateUserArgs struct {
    Name    string  `json:"name" jsonschema:"required"`
    Address Address `json:"address" jsonschema:"required"`
}
```

### Optional Fields

Fields without `required` tag are optional:

```go
type SearchArgs struct {
    Query    string `json:"query" jsonschema:"required"`
    Limit    *int   `json:"limit,omitempty"` // Optional, use pointer for zero value distinction
    Offset   int    `json:"offset"`           // Optional, defaults to 0
}
```

## Output Schemas

Define expected output structure for better type safety (MCP 2025-06-18):

```go
type WeatherOutput struct {
    Temperature float64 `json:"temperature"`
    Condition   string  `json:"condition"`
    Humidity    int     `json:"humidity"`
}

tool, _ := builder.NewTool("get_weather").
    Description("Get current weather").
    Handler(func(ctx context.Context, args LocationArgs) (*WeatherOutput, error) {
        // Implementation
        return &WeatherOutput{
            Temperature: 72.5,
            Condition:   "Sunny",
            Humidity:    45,
        }, nil
    }).
    OutputSchema(map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "temperature": map[string]interface{}{"type": "number"},
            "condition":   map[string]interface{}{"type": "string"},
            "humidity":    map[string]interface{}{"type": "integer"},
        },
        "required": []string{"temperature", "condition", "humidity"},
    }).
    Build()
```

## Tool Hints

Provide semantic hints about tool behavior (MCP 2025-03-26):

### ReadOnly Hint

Tool doesn't modify the environment:

```go
tool, _ := builder.NewTool("get_config").
    Description("Read configuration values").
    ReadOnlyHint(true).
    Handler(func(ctx context.Context, args GetConfigArgs) (interface{}, error) {
        // Read-only operation
    }).
    Build()
```

### Destructive Hint

Tool may perform destructive updates:

```go
tool, _ := builder.NewTool("delete_file").
    Description("Delete a file from the filesystem").
    DestructiveHint(true).
    Handler(func(ctx context.Context, args DeleteArgs) (bool, error) {
        // Destructive operation
    }).
    Build()
```

### Idempotent Hint

Repeated calls have no additional effect:

```go
tool, _ := builder.NewTool("create_user").
    Description("Create a new user").
    IdempotentHint(true).
    Handler(func(ctx context.Context, args CreateUserArgs) (*User, error) {
        // Idempotent: creating same user twice has same result
    }).
    Build()
```

### OpenWorld Hint

Tool may interact with external entities:

```go
tool, _ := builder.NewTool("send_email").
    Description("Send an email").
    OpenWorldHint(true).
    Handler(func(ctx context.Context, args EmailArgs) (bool, error) {
        // Interacts with external email service
    }).
    Build()
```

## Error Handling

### Standard Errors

Return descriptive errors from handlers:

```go
func (ctx context.Context, args DivideArgs) (float64, error) {
    if args.B == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return args.A / args.B, nil
}
```

### MCP Errors

Use MCP error codes for protocol-level errors:

```go
import "github.com/jmcarbo/fullmcp/mcp"

func (ctx context.Context, args GetUserArgs) (*User, error) {
    user, err := db.GetUser(args.ID)
    if err == sql.ErrNoRows {
        return nil, &mcp.Error{
            Code:    mcp.ErrorCodeResourceNotFound,
            Message: fmt.Sprintf("User %s not found", args.ID),
        }
    }
    return user, err
}
```

### Context Cancellation

Respect context cancellation:

```go
func (ctx context.Context, args ProcessArgs) (Result, error) {
    for i := 0; i < len(args.Items); i++ {
        select {
        case <-ctx.Done():
            return Result{}, ctx.Err()
        default:
            // Process item
        }
    }
    return result, nil
}
```

## Advanced Patterns

### Tool with External Dependencies

```go
type DatabaseTool struct {
    db *sql.DB
}

func (dt *DatabaseTool) GetUserTool() (*mcp.Tool, error) {
    return builder.NewTool("get_user").
        Description("Get user by ID").
        Handler(func(ctx context.Context, args GetUserArgs) (*User, error) {
            var user User
            err := dt.db.QueryRowContext(ctx,
                "SELECT id, name, email FROM users WHERE id = ?",
                args.ID,
            ).Scan(&user.ID, &user.Name, &user.Email)
            return &user, err
        }).
        Build()
}
```

### Dynamic Tool Registration

```go
// Register tools based on available plugins
for _, plugin := range plugins {
    tool, err := builder.NewTool(plugin.Name()).
        Description(plugin.Description()).
        Handler(plugin.Handler()).
        Build()

    if err != nil {
        return err
    }

    srv.AddTool(tool)
}
```

### Tool Middleware

Wrap handlers with common functionality:

```go
func withLogging(handler interface{}) interface{} {
    return func(ctx context.Context, args interface{}) (interface{}, error) {
        log.Printf("Calling tool with args: %+v", args)
        result, err := handler.(func(context.Context, interface{}) (interface{}, error))(ctx, args)
        log.Printf("Tool returned: %+v, error: %v", result, err)
        return result, err
    }
}

tool, _ := builder.NewTool("process").
    Handler(withLogging(actualHandler)).
    Build()
```

### Batch Operations

```go
type BatchProcessArgs struct {
    Items []string `json:"items" jsonschema:"required,minItems=1,maxItems=1000"`
}

type BatchResult struct {
    Processed int      `json:"processed"`
    Failed    int      `json:"failed"`
    Errors    []string `json:"errors,omitempty"`
}

tool, _ := builder.NewTool("batch_process").
    Description("Process multiple items in batch").
    Handler(func(ctx context.Context, args BatchProcessArgs) (*BatchResult, error) {
        result := &BatchResult{}

        for _, item := range args.Items {
            if err := processItem(ctx, item); err != nil {
                result.Failed++
                result.Errors = append(result.Errors, err.Error())
            } else {
                result.Processed++
            }
        }

        return result, nil
    }).
    Build()
```

## Best Practices

### Naming Conventions

✅ **Good:**
- `get_user`
- `create_post`
- `calculate_total`
- `search_files`

❌ **Bad:**
- `GetUser` (should be lowercase)
- `process` (too generic)
- `do_thing` (not descriptive)

### Input Validation

Always validate inputs:

```go
func (ctx context.Context, args CreateUserArgs) (*User, error) {
    // Validate email format
    if !strings.Contains(args.Email, "@") {
        return nil, fmt.Errorf("invalid email format")
    }

    // Validate age range
    if args.Age < 0 || args.Age > 150 {
        return nil, fmt.Errorf("age must be between 0 and 150")
    }

    // Create user...
}
```

### Error Messages

Provide clear, actionable error messages:

```go
// ✅ Good
return nil, fmt.Errorf("file %s not found in directory %s", filename, dir)

// ❌ Bad
return nil, fmt.Errorf("error")
```

### Context Usage

Always accept and use context:

```go
// ✅ Good
func (ctx context.Context, args Args) (Result, error) {
    return http.Get(ctx, url)
}

// ❌ Bad
func (ctx context.Context, args Args) (Result, error) {
    return http.Get(url) // Ignores context
}
```

### Documentation

Use clear descriptions:

```go
// ✅ Good
Description("Search for files matching the given pattern in the specified directory. Returns list of file paths.")

// ❌ Bad
Description("Search files")
```

### Type Safety

Prefer specific types over `interface{}`:

```go
// ✅ Good
type Result struct {
    Count int      `json:"count"`
    Items []string `json:"items"`
}

// ❌ Bad
type Result struct {
    Data interface{} `json:"data"`
}
```

### Performance

Consider performance for frequently-called tools:

```go
var schemaCache sync.Map

func getCachedSchema(key string) map[string]interface{} {
    if val, ok := schemaCache.Load(key); ok {
        return val.(map[string]interface{})
    }

    schema := generateSchema()
    schemaCache.Store(key, schema)
    return schema
}
```

## Testing Tools

### Unit Testing

```go
func TestAddTool(t *testing.T) {
    tool, err := builder.NewTool("add").
        Description("Add numbers").
        Handler(func(ctx context.Context, args AddArgs) (int, error) {
            return args.A + args.B, nil
        }).
        Build()

    require.NoError(t, err)
    assert.Equal(t, "add", tool.Name)
    assert.NotNil(t, tool.InputSchema)
}
```

### Integration Testing

```go
func TestToolExecution(t *testing.T) {
    srv := server.New("test")

    // Add tool
    addTool, _ := builder.NewTool("add").
        Handler(func(ctx context.Context, args AddArgs) (int, error) {
            return args.A + args.B, nil
        }).
        Build()

    srv.AddTool(addTool)

    // Call tool
    result, err := srv.ToolManager().CallTool(
        context.Background(),
        "add",
        json.RawMessage(`{"a": 5, "b": 3}`),
    )

    require.NoError(t, err)
    assert.Equal(t, 8, result)
}
```

## Examples

See [`examples/`](../examples/) directory for complete working examples:

- [basic-server](../examples/basic-server/): Simple calculator tools
- [advanced-server](../examples/advanced-server/): Complex tools with dependencies

## Related Documentation

- [Architecture Overview](./architecture.md)
- [Server API Reference](https://pkg.go.dev/github.com/jmcarbo/fullmcp/server)
- [Builder API Reference](https://pkg.go.dev/github.com/jmcarbo/fullmcp/builder)
