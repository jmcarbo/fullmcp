# Using Prompts

Prompts are reusable message templates that can be parameterized and rendered into conversation messages. They enable consistent interaction patterns and help build structured conversations with AI models.

## Table of Contents

- [Overview](#overview)
- [Creating Prompts](#creating-prompts)
- [Arguments](#arguments)
- [Message Rendering](#message-rendering)
- [Content Types](#content-types)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

Prompts in MCP serve as templates for generating structured messages:
- Parameterized message generation
- Reusable conversation patterns
- Type-safe argument handling
- Multi-message support
- Rich content support (text, images, resources)

**Use Cases:**
- Standardized greetings or introductions
- Complex multi-step instructions
- Role-based conversation starters
- Domain-specific question templates
- Code review or analysis templates

## Creating Prompts

### Basic Prompt

```go
import (
    "context"
    "fmt"
    "github.com/jmcarbo/fullmcp/builder"
    "github.com/jmcarbo/fullmcp/mcp"
    "github.com/jmcarbo/fullmcp/server"
)

func main() {
    srv := server.New("prompt-server")

    greetingPrompt := builder.NewPrompt("greeting").
        Description("Generate a personalized greeting").
        Argument("name", "Person's name", true).
        Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
            name := args["name"].(string)

            return []*mcp.PromptMessage{{
                Role: "user",
                Content: []mcp.Content{
                    &mcp.TextContent{
                        Type: "text",
                        Text: fmt.Sprintf("Hello, %s! Welcome to our service.", name),
                    },
                },
            }}, nil
        }).
        Build()

    srv.AddPrompt(greetingPrompt)
}
```

### Builder Methods

#### Name (set via constructor)

```go
func NewPrompt(name string) *PromptBuilder
```

Unique prompt identifier.

#### Description

```go
func (pb *PromptBuilder) Description(desc string) *PromptBuilder
```

Describes what the prompt does:

```go
prompt := builder.NewPrompt("code_review").
    Description("Generate a code review request for a given file").
    // ...
```

#### Title

```go
func (pb *PromptBuilder) Title(title string) *PromptBuilder
```

Human-friendly display name (MCP 2025-06-18):

```go
prompt := builder.NewPrompt("analyze_code").
    Title("Code Analysis Request").
    Description("Request detailed code analysis").
    // ...
```

## Arguments

### Adding Arguments

```go
func (pb *PromptBuilder) Argument(name, description string, required bool) *PromptBuilder
```

Define prompt parameters:

```go
prompt := builder.NewPrompt("summarize").
    Description("Generate a summary request").
    Argument("text", "Text to summarize", true).
    Argument("maxLength", "Maximum summary length", false).
    Argument("style", "Summary style (brief, detailed, technical)", false).
    // ...
```

### Accessing Arguments

Arguments are passed to the renderer as `map[string]interface{}`:

```go
Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
    // Required arguments
    text := args["text"].(string)

    // Optional arguments with defaults
    maxLength := 100
    if ml, ok := args["maxLength"].(float64); ok {
        maxLength = int(ml)
    }

    style := "brief"
    if s, ok := args["style"].(string); ok {
        style = s
    }

    // Use arguments...
})
```

### Type Assertions

Handle different argument types safely:

```go
Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
    // String
    name, ok := args["name"].(string)
    if !ok {
        return nil, fmt.Errorf("name must be a string")
    }

    // Number (JSON numbers are float64)
    age, ok := args["age"].(float64)
    if !ok {
        return nil, fmt.Errorf("age must be a number")
    }

    // Boolean
    active, ok := args["active"].(bool)
    if !ok {
        active = true // default
    }

    // Array
    tags, ok := args["tags"].([]interface{})
    if ok {
        stringTags := make([]string, len(tags))
        for i, tag := range tags {
            stringTags[i] = tag.(string)
        }
    }

    // ...
})
```

## Message Rendering

### Renderer Function

```go
func (pb *PromptBuilder) Renderer(
    fn func(context.Context, map[string]interface{}) ([]*mcp.PromptMessage, error)
) *PromptBuilder
```

The renderer function generates messages based on arguments.

### Single Message

```go
Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
    topic := args["topic"].(string)

    return []*mcp.PromptMessage{{
        Role: "user",
        Content: []mcp.Content{
            &mcp.TextContent{
                Type: "text",
                Text: fmt.Sprintf("Please explain %s in detail.", topic),
            },
        },
    }}, nil
})
```

### Multiple Messages

Create multi-turn conversation templates:

```go
Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
    language := args["language"].(string)
    task := args["task"].(string)

    return []*mcp.PromptMessage{
        {
            Role: "system",
            Content: []mcp.Content{
                &mcp.TextContent{
                    Type: "text",
                    Text: fmt.Sprintf("You are an expert %s programmer.", language),
                },
            },
        },
        {
            Role: "user",
            Content: []mcp.Content{
                &mcp.TextContent{
                    Type: "text",
                    Text: fmt.Sprintf("Help me %s", task),
                },
            },
        },
    }, nil
})
```

### Message Roles

Standard message roles:
- `"system"`: System-level instructions
- `"user"`: User messages
- `"assistant"`: Assistant responses (for few-shot examples)

## Content Types

### Text Content

```go
&mcp.TextContent{
    Type: "text",
    Text: "Your text here",
}
```

### Image Content

```go
&mcp.ImageContent{
    Type:     "image",
    Data:     base64EncodedImage,
    MimeType: "image/png",
}
```

Example with image:

```go
Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
    imagePath := args["imagePath"].(string)

    imageData, err := os.ReadFile(imagePath)
    if err != nil {
        return nil, err
    }

    encoded := base64.StdEncoding.EncodeToString(imageData)

    return []*mcp.PromptMessage{{
        Role: "user",
        Content: []mcp.Content{
            &mcp.TextContent{
                Type: "text",
                Text: "Please analyze this image:",
            },
            &mcp.ImageContent{
                Type:     "image",
                Data:     encoded,
                MimeType: "image/png",
            },
        },
    }}, nil
})
```

### Resource Content

Reference existing resources:

```go
&mcp.ResourceContent{
    Type:     "resource",
    URI:      "file:///config/settings.json",
    MimeType: "application/json",
}
```

### Mixed Content

Combine multiple content types:

```go
return []*mcp.PromptMessage{{
    Role: "user",
    Content: []mcp.Content{
        &mcp.TextContent{
            Type: "text",
            Text: "Review this configuration file:",
        },
        &mcp.ResourceContent{
            Type:     "resource",
            URI:      "config://app",
            MimeType: "application/json",
        },
        &mcp.TextContent{
            Type: "text",
            Text: "Look for security issues.",
        },
    },
}}, nil
```

## Best Practices

### Naming Conventions

✅ **Good:**
- `code_review`
- `summarize_text`
- `analyze_data`
- `generate_report`

❌ **Bad:**
- `CodeReview` (use lowercase with underscores)
- `prompt1` (not descriptive)
- `do_thing` (too generic)

### Argument Validation

Validate arguments before use:

```go
Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
    // Check required arguments
    text, ok := args["text"].(string)
    if !ok || text == "" {
        return nil, fmt.Errorf("text argument is required and must be non-empty")
    }

    // Validate argument values
    if len(text) > 10000 {
        return nil, fmt.Errorf("text too long (max 10000 characters)")
    }

    // Validate enum values
    style, ok := args["style"].(string)
    if ok {
        validStyles := map[string]bool{
            "brief": true, "detailed": true, "technical": true,
        }
        if !validStyles[style] {
            return nil, fmt.Errorf("invalid style: %s", style)
        }
    }

    // ...
})
```

### Error Handling

Return clear error messages:

```go
Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
    filePath, ok := args["file"].(string)
    if !ok {
        return nil, fmt.Errorf("file argument must be a string")
    }

    content, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
    }

    // ...
})
```

### Context Usage

Respect context cancellation:

```go
Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        // Render prompt
    }
})
```

### Template Reusability

Design prompts to be reusable across different scenarios:

```go
// ✅ Good: Generic and reusable
builder.NewPrompt("analyze").
    Argument("subject", "What to analyze", true).
    Argument("focus", "Specific focus area", false).
    Argument("depth", "Analysis depth (shallow, moderate, deep)", false)

// ❌ Bad: Too specific
builder.NewPrompt("analyze_user_123_profile")
```

## Examples

### Code Review Prompt

```go
codeReviewPrompt := builder.NewPrompt("code_review").
    Title("Code Review Request").
    Description("Generate a code review request").
    Argument("file", "File path to review", true).
    Argument("focus", "Areas to focus on (security, performance, style)", false).
    Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
        filePath := args["file"].(string)

        content, err := os.ReadFile(filePath)
        if err != nil {
            return nil, err
        }

        focus := "general code quality"
        if f, ok := args["focus"].(string); ok {
            focus = f
        }

        return []*mcp.PromptMessage{
            {
                Role: "system",
                Content: []mcp.Content{
                    &mcp.TextContent{
                        Type: "text",
                        Text: "You are an expert code reviewer.",
                    },
                },
            },
            {
                Role: "user",
                Content: []mcp.Content{
                    &mcp.TextContent{
                        Type: "text",
                        Text: fmt.Sprintf(
                            "Please review this code with focus on %s:\n\n```\n%s\n```",
                            focus,
                            string(content),
                        ),
                    },
                },
            },
        }, nil
    }).
    Build()
```

### Data Analysis Prompt

```go
analysisPrompt := builder.NewPrompt("analyze_data").
    Description("Generate data analysis request").
    Argument("dataset", "Dataset URI", true).
    Argument("questions", "Specific questions to answer", false).
    Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
        datasetURI := args["dataset"].(string)

        message := fmt.Sprintf("Please analyze the dataset at %s", datasetURI)

        if questions, ok := args["questions"].([]interface{}); ok {
            message += "\n\nSpecifically answer these questions:\n"
            for i, q := range questions {
                message += fmt.Sprintf("%d. %s\n", i+1, q.(string))
            }
        }

        return []*mcp.PromptMessage{{
            Role: "user",
            Content: []mcp.Content{
                &mcp.TextContent{
                    Type: "text",
                    Text: message,
                },
                &mcp.ResourceContent{
                    Type: "resource",
                    URI:  datasetURI,
                },
            },
        }}, nil
    }).
    Build()
```

### Multi-Turn Conversation

```go
tutorPrompt := builder.NewPrompt("tutor").
    Description("Start a tutoring session").
    Argument("subject", "Subject to teach", true).
    Argument("level", "Student level (beginner, intermediate, advanced)", true).
    Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
        subject := args["subject"].(string)
        level := args["level"].(string)

        return []*mcp.PromptMessage{
            {
                Role: "system",
                Content: []mcp.Content{
                    &mcp.TextContent{
                        Type: "text",
                        Text: fmt.Sprintf(
                            "You are a patient tutor teaching %s to %s level students.",
                            subject, level,
                        ),
                    },
                },
            },
            {
                Role: "user",
                Content: []mcp.Content{
                    &mcp.TextContent{
                        Type: "text",
                        Text: "I'm ready to learn. Where should we start?",
                    },
                },
            },
            {
                Role: "assistant",
                Content: []mcp.Content{
                    &mcp.TextContent{
                        Type: "text",
                        Text: fmt.Sprintf(
                            "Great! Let's start with the fundamentals of %s. "+
                                "I'll explain concepts clearly and check your understanding as we go.",
                            subject,
                        ),
                    },
                },
            },
            {
                Role: "user",
                Content: []mcp.Content{
                    &mcp.TextContent{
                        Type: "text",
                        Text: "Please begin with the basics.",
                    },
                },
            },
        }, nil
    }).
    Build()
```

## Testing Prompts

### Unit Testing

```go
func TestPrompt(t *testing.T) {
    prompt := builder.NewPrompt("test").
        Description("Test prompt").
        Argument("name", "Name", true).
        Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
            name := args["name"].(string)
            return []*mcp.PromptMessage{{
                Role: "user",
                Content: []mcp.Content{
                    &mcp.TextContent{
                        Type: "text",
                        Text: fmt.Sprintf("Hello, %s", name),
                    },
                },
            }}, nil
        }).
        Build()

    require.NotNil(t, prompt)
    assert.Equal(t, "test", prompt.Name)
    assert.Len(t, prompt.Arguments, 1)
}
```

### Integration Testing

```go
func TestPromptRendering(t *testing.T) {
    srv := server.New("test")

    prompt := builder.NewPrompt("greeting").
        Argument("name", "Name", true).
        Renderer(func(ctx context.Context, args map[string]interface{}) ([]*mcp.PromptMessage, error) {
            return []*mcp.PromptMessage{{
                Role: "user",
                Content: []mcp.Content{
                    &mcp.TextContent{
                        Type: "text",
                        Text: fmt.Sprintf("Hello, %s", args["name"]),
                    },
                },
            }}, nil
        }).
        Build()

    srv.AddPrompt(prompt)

    messages, err := srv.PromptManager().GetPrompt(
        context.Background(),
        "greeting",
        map[string]interface{}{"name": "Alice"},
    )

    require.NoError(t, err)
    require.Len(t, messages, 1)
    assert.Equal(t, "user", messages[0].Role)
}
```

## Client Usage

```go
import "github.com/jmcarbo/fullmcp/client"

// List available prompts
prompts, err := client.ListPrompts(ctx)

// Get rendered prompt
messages, err := client.GetPrompt(ctx, "code_review", map[string]interface{}{
    "file":  "/path/to/file.go",
    "focus": "security",
})

// Use messages in conversation...
```

## Related Documentation

- [Architecture Overview](./architecture.md)
- [Building Tools](./tools.md)
- [Managing Resources](./resources.md)
- [Prompt API Reference](https://pkg.go.dev/github.com/jmcarbo/fullmcp/server#PromptManager)
