// Package main demonstrates completion/complete for argument autocompletion in MCP
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmcarbo/fullmcp/mcp"
	"github.com/jmcarbo/fullmcp/server"
)

func setupServerWithCompletion() *server.Server {
	fmt.Println("💡 Server with Completion Support")
	fmt.Println("==================================")
	fmt.Println()

	srv := server.New("completion-demo", server.WithCompletion())
	fmt.Println("✓ Server created with completion support")
	fmt.Println()
	return srv
}

func registerCompletionHandlers(srv *server.Server) {
	fmt.Println("📝 Register Completion Handlers")
	fmt.Println("================================")
	fmt.Println()

	srv.RegisterPromptCompletion("code_review", func(_ context.Context, _ mcp.CompletionRef, arg mcp.CompletionArgument) ([]string, error) {
		if arg.Name == "language" {
			languages := []string{"Go", "Python", "JavaScript", "TypeScript", "Rust", "Java"}
			if arg.Value != "" {
				var filtered []string
				for _, lang := range languages {
					if strings.HasPrefix(strings.ToLower(lang), strings.ToLower(arg.Value)) {
						filtered = append(filtered, lang)
					}
				}
				return filtered, nil
			}
			return languages, nil
		}
		return []string{}, nil
	})

	fmt.Println("✓ Registered prompt completion for 'code_review'")
	fmt.Println()
}

func showCompletionRequestStructure() {
	fmt.Println("📋 Completion Request Structure")
	fmt.Println("================================")
	fmt.Println()

	request := mcp.CompleteRequest{
		Ref: mcp.CompletionRef{
			Type: "ref/prompt",
			Name: "code_review",
		},
		Argument: mcp.CompletionArgument{
			Name:  "language",
			Value: "Go",
		},
	}

	fmt.Printf("   Ref Type:      %s\n", request.Ref.Type)
	fmt.Printf("   Ref Name:      %s\n", request.Ref.Name)
	fmt.Printf("   Argument Name: %s\n", request.Argument.Name)
	fmt.Printf("   Partial Value: %s\n", request.Argument.Value)
	fmt.Println()

	// Example 4: Reference types
	fmt.Println("🔖 Reference Types")
	fmt.Println("==================")
	fmt.Println()

	refTypes := []struct {
		refType string
		desc    string
		example string
	}{
		{"ref/prompt", "Completion for prompt arguments", `{"type": "ref/prompt", "name": "code_review"}`},
		{"ref/resource", "Completion for resource URIs", `{"type": "ref/resource", "name": "file:///"}`},
	}

	for i, rt := range refTypes {
		fmt.Printf("   %d. %s\n", i+1, rt.refType)
		fmt.Printf("      %s\n", rt.desc)
		fmt.Printf("      Example: %s\n", rt.example)
		fmt.Println()
	}

	// Example 5: Use cases
	fmt.Println("💼 Common Use Cases")
	fmt.Println("===================")
	fmt.Println()

	useCases := []struct {
		title       string
		description string
		example     string
	}{
		{
			"Language Selection",
			"Autocomplete programming languages",
			"Go, Python, JavaScript...",
		},
		{
			"File Paths",
			"Complete file paths and directories",
			"/home/user/Documents/...",
		},
		{
			"Configuration Keys",
			"Suggest valid config options",
			"max_workers, timeout, debug...",
		},
		{
			"Resource URIs",
			"Complete resource identifiers",
			"config://app, db://users...",
		},
		{
			"Enum Values",
			"Suggest from predefined sets",
			"INFO, WARN, ERROR...",
		},
		{
			"User Names",
			"Autocomplete user identifiers",
			"@alice, @bob, @charlie...",
		},
	}

	for i, uc := range useCases {
		fmt.Printf("   %d. %s\n", i+1, uc.title)
		fmt.Printf("      %s\n", uc.description)
		fmt.Printf("      Example: %s\n", uc.example)
		fmt.Println()
	}

	// Example 6: Filtering completions
	fmt.Println("🔍 Filtering Completions")
	fmt.Println("========================")
	fmt.Println()

	fmt.Println("User types partial value, server filters results:")
	fmt.Println()

	allLanguages := []string{"Go", "Python", "JavaScript", "TypeScript", "Rust", "Java"}
	testInputs := []string{"", "J", "Ja", "Py", "Ru"}

	for _, input := range testInputs {
		fmt.Printf("   Input: \"%s\"\n", input)
		var matches []string
		for _, lang := range allLanguages {
			if input == "" || strings.HasPrefix(strings.ToLower(lang), strings.ToLower(input)) {
				matches = append(matches, lang)
			}
		}
		fmt.Printf("   Suggestions: %v\n", matches)
		fmt.Println()
	}

	// Example 7: Client-side usage
	fmt.Println("🔌 Client-Side Usage")
	fmt.Println("====================")
	fmt.Println()

	fmt.Println("Request completions:")
	fmt.Print(`
  ref := mcp.CompletionRef{
    Type: "ref/prompt",
    Name: "code_review",
  }

  arg := mcp.CompletionArgument{
    Name:  "language",
    Value: "Ja", // User typed "Ja"
  }

  suggestions, err := client.GetCompletion(ctx, ref, arg)
  // suggestions: ["Java", "JavaScript"]
`)
	fmt.Println()

	// Example 8: Server implementation example
	fmt.Println("🖥️  Server Implementation")
	fmt.Println("=========================")
	fmt.Println()

	fmt.Println("File path completion example:")
	fmt.Print(`
  srv.RegisterResourceCompletion("file:///",
    func(ctx context.Context, ref mcp.CompletionRef, arg mcp.CompletionArgument) ([]string, error) {
      if arg.Name == "path" {
        // List directory contents
        dir := filepath.Dir(arg.Value)
        entries, _ := os.ReadDir(dir)

        var paths []string
        for _, entry := range entries {
          fullPath := filepath.Join(dir, entry.Name())
          if strings.HasPrefix(fullPath, arg.Value) {
            paths = append(paths, fullPath)
          }
        }
        return paths, nil
      }
      return []string{}, nil
    },
  )
`)

	// Example 9: Rich completions
	fmt.Println("🎨 Rich Completions (Advanced)")
	fmt.Println("==============================")
	fmt.Println()

	fmt.Println("CompletionValue with metadata:")
	richCompletions := []mcp.CompletionValue{
		{
			Value:  "Go",
			Label:  "Go 1.21",
			Detail: "Statically typed, compiled language",
			Data: map[string]interface{}{
				"version": "1.21",
				"type":    "compiled",
			},
		},
		{
			Value:  "Python",
			Label:  "Python 3.11",
			Detail: "High-level, interpreted language",
			Data: map[string]interface{}{
				"version": "3.11",
				"type":    "interpreted",
			},
		},
	}

	for i, comp := range richCompletions {
		fmt.Printf("   %d. %s\n", i+1, comp.Value)
		fmt.Printf("      Label:  %s\n", comp.Label)
		fmt.Printf("      Detail: %s\n", comp.Detail)
		fmt.Printf("      Data:   %v\n", comp.Data)
		fmt.Println()
	}

	// Example 10: Best practices
	fmt.Println("📋 Best Practices")
	fmt.Println("=================")
	fmt.Println()

	fmt.Println("For Servers:")
	fmt.Println("   ✓ Filter results based on partial input")
	fmt.Println("   ✓ Return most relevant suggestions first")
	fmt.Println("   ✓ Limit number of suggestions (e.g., top 10)")
	fmt.Println("   ✓ Use case-insensitive matching")
	fmt.Println("   ✓ Consider fuzzy matching for better UX")
	fmt.Println()

	fmt.Println("For Clients:")
	fmt.Println("   ✓ Show completions as user types")
	fmt.Println("   ✓ Update suggestions with each keystroke")
	fmt.Println("   ✓ Display in dropdown or popup menu")
	fmt.Println("   ✓ Handle keyboard navigation")
	fmt.Println("   ✓ Cache results for better performance")
	fmt.Println()

	// Example 11: Protocol flow
	fmt.Println("🔄 Protocol Flow")
	fmt.Println("================")
	fmt.Println()

	fmt.Println("1. User types partial value in UI")
	fmt.Println("2. Client sends completion/complete request")
	fmt.Println("3. Server filters and returns suggestions")
	fmt.Println("4. Client displays suggestions to user")
	fmt.Println("5. User selects completion or continues typing")
	fmt.Println()

	fmt.Println("✨ Completion demonstration complete!")
	fmt.Println()
	fmt.Println("Note:")
	fmt.Println("  - Enable completion with WithCompletion()")
	fmt.Println("  - Register handlers for prompts and resources")
	fmt.Println("  - Filter suggestions based on partial input")
	fmt.Println("  - Return empty array if no handler registered")
	fmt.Println("  - Provides IDE-like autocomplete experience")
}

func main() {
	fmt.Println("MCP Completion Example")
	fmt.Println("======================")
	fmt.Println()

	srv := setupServerWithCompletion()
	registerCompletionHandlers(srv)
	_ = srv

	showCompletionRequestStructure()
}
