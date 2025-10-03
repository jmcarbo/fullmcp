# Typed Content Support

As of this update, FullMCP supports proper typed content handling across **Tools**, **Resources**, and **Prompts**.

## Tools - Typed Content Support

## Overview

Previously, all tool responses were converted to text using `fmt.Sprintf("%v", result)`, which meant:
- ❌ Tool handlers couldn't return typed content (images, audio, etc.)
- ❌ MIME types were not preserved
- ❌ Binary data got mangled by string conversion
- ❌ Structured content lost its type information

Now, tools can return:
- ✅ Text content (`mcp.TextContent`, `string`)
- ✅ Image content (`mcp.ImageContent`) with MIME types
- ✅ Audio content (`mcp.AudioContent`) with MIME types
- ✅ Resource content (`mcp.ResourceContent`, `mcp.ResourceLinkContent`)
- ✅ Multiple content blocks (`[]mcp.Content`)
- ✅ Structured data (structs, maps) - auto-converted to JSON
- ✅ Backward compatibility with simple types

## Return Types

### Direct MCP Content Types

Return MCP content types directly for full control:

```go
// TextContent
builder.NewTool("echo").
    Handler(func(_ context.Context, input struct{ Text string }) (mcp.TextContent, error) {
        return mcp.TextContent{
            Type: "text",
            Text: input.Text,
        }, nil
    })

// ImageContent
builder.NewTool("generate_image").
    Handler(func(_ context.Context, input ImageInput) (mcp.ImageContent, error) {
        return mcp.ImageContent{
            Type:     "image",
            Data:     base64EncodedData,
            MimeType: "image/png",
        }, nil
    })

// AudioContent
builder.NewTool("generate_audio").
    Handler(func(_ context.Context, input AudioInput) (mcp.AudioContent, error) {
        return mcp.AudioContent{
            Type:     "audio",
            Data:     base64EncodedData,
            MimeType: "audio/wav",
        }, nil
    })

// ResourceContent
builder.NewTool("get_resource").
    Handler(func(_ context.Context, input struct{}) (mcp.ResourceContent, error) {
        return mcp.ResourceContent{
            Type:     "resource",
            URI:      "file:///data.json",
            MimeType: "application/json",
            Text:     jsonData,
        }, nil
    })
```

### Multiple Content Blocks

Return a slice of content for multiple blocks:

```go
builder.NewTool("analyze").
    Handler(func(_ context.Context, input DataInput) ([]mcp.Content, error) {
        return []mcp.Content{
            mcp.TextContent{Type: "text", Text: "Analysis:"},
            mcp.TextContent{Type: "text", Text: "Result: ..."},
            mcp.ResourceContent{
                Type:     "resource",
                URI:      "data://processed",
                MimeType: "application/json",
                Text:     processedData,
            },
        }, nil
    })
```

### Simple Types (Backward Compatible)

Simple types are automatically converted:

```go
// String → TextContent
builder.NewTool("hello").
    Handler(func(_ context.Context, input struct{}) (string, error) {
        return "Hello, World!", nil
    })

// Integer → TextContent (JSON)
builder.NewTool("count").
    Handler(func(_ context.Context, input struct{}) (int, error) {
        return 42, nil
    })

// Struct → TextContent (JSON)
builder.NewTool("get_data").
    Handler(func(_ context.Context, input struct{}) (MyStruct, error) {
        return MyStruct{Name: "test", Value: 123}, nil
    })

// Map → TextContent (JSON)
builder.NewTool("get_config").
    Handler(func(_ context.Context, input struct{}) (map[string]interface{}, error) {
        return map[string]interface{}{
            "enabled": true,
            "count":   10,
        }, nil
    })
```

## Conversion Logic

The `convertToContent` function in `server/server.go` handles conversion:

1. **nil** → Empty TextContent
2. **[]mcp.Content** → Used directly
3. **mcp.Content types** → Wrapped in slice
4. **string** → TextContent
5. **[]byte** → TextContent (as string)
6. **Other types** → JSON marshaled to TextContent

## Example

See `examples/typed-content/main.go` for a complete working example demonstrating:
- TextContent, ImageContent, AudioContent
- Multiple content blocks
- Backward-compatible simple types

## Testing

Comprehensive tests are available in `server/content_conversion_test.go` covering all conversion scenarios.

## Migration

Existing tools require no changes - they continue to work exactly as before. To take advantage of typed content:

1. Change your tool handler's return type from a simple type to an MCP Content type
2. Return the appropriate content with MIME types
3. Test with clients that support rich content

## Resources - MIME Type Preservation

Previously, all resources were returned with hardcoded `"text/plain"` MIME type. Now resources preserve their actual MIME types.

### Problem Fixed

```go
// BEFORE - hardcoded MIME type
contents := []map[string]interface{}{
    {
        "uri":      params.URI,
        "mimeType": "text/plain",  // ❌ Always text/plain!
        "text":     string(data),
    },
}
```

### Solution

Resources now use the MIME type specified in their handler:

```go
srv.AddResource(&server.ResourceHandler{
    URI:         "config://app.json",
    Name:        "app-config",
    Description: "Application configuration",
    MimeType:    "application/json",  // ✅ Preserved!
    Reader: func(_ context.Context) ([]byte, error) {
        return []byte(`{"debug": true}`), nil
    },
})
```

The server now:
1. Retrieves resource metadata along with data
2. Returns the actual MIME type from the resource handler
3. Defaults to "text/plain" if no MIME type is specified
4. Maintains backward compatibility with the legacy `Read()` method

### New API

```go
// New method with metadata
content, err := rm.ReadWithMetadata(ctx, uri)
// content.Data     []byte
// content.MimeType string
// content.URI      string

// Legacy method still works
data, err := rm.Read(ctx, uri)
```

## Prompts - Already Correct

Prompts already support typed content correctly through `[]mcp.Content` in `PromptMessage`. No changes needed.

```go
srv.AddPrompt(&server.PromptHandler{
    Name: "greeting",
    Renderer: func(_ context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
        return []*mcp.PromptMessage{
            {
                Role: "user",
                Content: []mcp.Content{
                    mcp.TextContent{Type: "text", Text: "Hello!"},
                    mcp.ImageContent{Type: "image", Data: base64Data, MimeType: "image/png"},
                },
            },
        }, nil
    },
})
```

## Benefits

- **Type Safety**: Return exactly what you mean
- **MIME Type Preservation**: Tools, resources, images, and audio maintain their types
- **Multiple Content Blocks**: Return complex responses
- **Backward Compatible**: Existing code continues to work
- **Better JSON Representation**: Structs/maps are properly JSON-encoded instead of using `fmt.Sprintf`
