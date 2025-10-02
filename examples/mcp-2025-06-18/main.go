// Package main demonstrates MCP 2025-06-18 specification features
package main

import (
	"fmt"
)

func main() {
	fmt.Println("MCP 2025-06-18 Specification Features")
	fmt.Println("=====================================")
	fmt.Println()

	// Example 1: What's New in 2025-06-18
	fmt.Println("💡 What's New in MCP 2025-06-18")
	fmt.Println("================================")
	fmt.Println()
	fmt.Println("Key Changes from 2025-03-26:")
	fmt.Println("  ❌ REMOVED: JSON-RPC batching")
	fmt.Println("  ✅ ADDED: Tool output schemas")
	fmt.Println("  ✅ ADDED: Elicitation capability")
	fmt.Println("  ✅ ADDED: _meta fields on types")
	fmt.Println("  ✅ ADDED: title fields for display names")
	fmt.Println("  ✅ ADDED: Resource links in tool results")
	fmt.Println("  ✅ REQUIRED: MCP-Protocol-Version header")
	fmt.Println()

	// Example 2: Tool Output Schemas
	fmt.Println("🔧 Tool Output Schemas (NEW)")
	fmt.Println("============================")
	fmt.Println()
	fmt.Println("Tools can now specify expected output structure:")
	fmt.Print(`
  tool := Tool{
      Name: "analyze_code",
      InputSchema: {...},
      OutputSchema: {  // 2025-06-18
          "type": "object",
          "properties": {
              "complexity": {"type": "number"},
              "issues": {"type": "array"},
              "score": {"type": "number", "minimum": 0, "maximum": 100}
          },
          "required": ["complexity", "issues", "score"]
      }
  }
`)
	fmt.Println("Benefits:")
	fmt.Println("  • Clients know output structure before calling")
	fmt.Println("  • LLMs can better understand tool results")
	fmt.Println("  • Enables better validation and type checking")
	fmt.Println()

	// Example 3: Using Builder with Output Schema
	fmt.Println("🏗️  Builder with Output Schema")
	fmt.Println("===============================")
	fmt.Println()
	fmt.Println("Automatic schema generation from Go types:")
	fmt.Print(`
  type AnalysisResult struct {
      Complexity int      `+"`json:\"complexity\"`"+`
      Issues     []string `+"`json:\"issues\"`"+`
      Score      float64  `+"`json:\"score\"`"+`
  }

  tool := builder.NewTool("analyze_code").
      Handler(func(ctx context.Context, input CodeInput) (AnalysisResult, error) {
          // Tool implementation
      }).
      OutputSchemaFromType(AnalysisResult{}).  // Auto-generates schema!
      Build()
`)
	fmt.Println()

	// Example 4: _meta Fields
	fmt.Println("📋 _meta Fields (NEW)")
	fmt.Println("=====================")
	fmt.Println()
	fmt.Println("Resources, prompts, and templates can include metadata:")
	fmt.Print(`
  resource := Resource{
      URI:  "file:///project/data.json",
      Name: "Project Data",
      Title: "Project Configuration Data",  // Human-friendly
      _meta: {
          "lastModified": "2025-06-18T10:30:00Z",
          "audience": ["user", "assistant"],
          "priority": 0.8,
          "author": "dev-team",
          "version": "2.1.0"
      }
  }
`)
	fmt.Println("Use Cases:")
	fmt.Println("  • Track resource versions")
	fmt.Println("  • Audience targeting (user/assistant)")
	fmt.Println("  • Priority ordering")
	fmt.Println("  • Custom application metadata")
	fmt.Println()

	// Example 5: Title Fields
	fmt.Println("✨ Title Fields (NEW)")
	fmt.Println("====================")
	fmt.Println()
	fmt.Println("Display-friendly names for resources and prompts:")
	fmt.Print(`
  prompt := Prompt{
      Name:  "code_review",           // Technical ID
      Title: "Code Review Assistant",  // Display name
      Description: "Reviews code for quality and best practices"
  }

  resource := Resource{
      URI:   "file:///docs/api.md",   // Technical URI
      Name:  "api_docs",               // Technical ID
      Title: "API Documentation",      // Display name
  }
`)
	fmt.Println("Benefits:")
	fmt.Println("  • Better UX in client applications")
	fmt.Println("  • Separate technical IDs from display text")
	fmt.Println("  • Easier internationalization")
	fmt.Println()

	// Example 6: Resource Links in Tool Results
	fmt.Println("🔗 Resource Links in Tool Results (NEW)")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("Tools can return resource links:")
	fmt.Print(`
  result := []Content{
      TextContent{
          Type: "text",
          Text: "Analysis complete. See full report.",
      },
      ResourceLinkContent{  // 2025-06-18
          Type: "resource",
          Resource: Resource{
              URI:  "file:///reports/analysis.pdf",
              Name: "Analysis Report",
              Title: "Detailed Analysis Report",
              MimeType: "application/pdf",
          },
          Annotations: {
              "generated": "2025-06-18T10:30:00Z",
              "size": 2048000,
          },
      },
  }
`)
	fmt.Println("Use Cases:")
	fmt.Println("  • Link to generated reports")
	fmt.Println("  • Reference analysis results")
	fmt.Println("  • Point to created files")
	fmt.Println("  • Attach supplementary data")
	fmt.Println()

	// Example 7: Elicitation Capability
	fmt.Println("💬 Elicitation Capability (NEW)")
	fmt.Println("================================")
	fmt.Println()
	fmt.Println("Servers can request structured user input:")
	fmt.Print(`
  // Server requests user information
  request := ElicitationRequest{
      Description: "Please provide API configuration",
      Schema: {
          "type": "object",
          "properties": {
              "api_key": {
                  "type": "string",
                  "description": "Your API key"
              },
              "region": {
                  "type": "string",
                  "enum": ["us-east", "us-west", "eu"]
              }
          },
          "required": ["api_key"]
      }
  }

  // User can: Accept, Decline, or Cancel
  response := ElicitationResponse{
      Action: "accept",
      Data: {
          "api_key": "sk-...",
          "region": "us-east"
      }
  }
`)
	fmt.Println("Features:")
	fmt.Println("  • Structured data collection")
	fmt.Println("  • JSON Schema validation")
	fmt.Println("  • User-controlled information sharing")
	fmt.Println("  • Three response actions: accept/decline/cancel")
	fmt.Println()

	// Example 8: Elicitation Security
	fmt.Println("🔒 Elicitation Security")
	fmt.Println("=======================")
	fmt.Println()
	fmt.Println("Security Guidelines:")
	fmt.Println("  • Servers MUST NOT request sensitive info without justification")
	fmt.Println("  • Clients SHOULD show clear server identification")
	fmt.Println("  • Clients SHOULD allow response review before submission")
	fmt.Println("  • Clients SHOULD offer decline/cancel options")
	fmt.Println("  • Clients SHOULD implement rate limiting")
	fmt.Println()
	fmt.Println("Schema Restrictions:")
	fmt.Println("  • Primitive types only: string, number, boolean, enum")
	fmt.Println("  • Flat object structures (no nesting)")
	fmt.Println("  • Validation constraints: min/max, formats, patterns")
	fmt.Println()

	// Example 9: Protocol Version Header
	fmt.Println("🏷️  MCP-Protocol-Version Header (REQUIRED)")
	fmt.Println("==========================================")
	fmt.Println()
	fmt.Println("All HTTP-based transports must include version header:")
	fmt.Print(`
  // Client request
  POST /mcp HTTP/1.1
  Content-Type: application/json
  MCP-Protocol-Version: 2025-06-18

  {...}

  // Server response
  HTTP/1.1 200 OK
  Content-Type: application/json
  MCP-Protocol-Version: 2025-06-18

  {...}
`)
	fmt.Println("Version Negotiation:")
	fmt.Println("  1. Client sends latest supported version")
	fmt.Println("  2. Server responds with same or alternative version")
	fmt.Println("  3. Client disconnects if incompatible")
	fmt.Println()

	// Example 10: Breaking Changes
	fmt.Println("⚠️  Breaking Changes from 2025-03-26")
	fmt.Println("====================================")
	fmt.Println()
	fmt.Println("JSON-RPC Batching Removed:")
	fmt.Print(`
  // Before (2025-03-26):
  [
      {"jsonrpc": "2.0", "method": "tools/list", "id": 1},
      {"jsonrpc": "2.0", "method": "resources/list", "id": 2}
  ]

  // After (2025-06-18):
  // Send individual requests only
  {"jsonrpc": "2.0", "method": "tools/list", "id": 1}
`)
	fmt.Println("Rationale:")
	fmt.Println("  • Simplifies protocol")
	fmt.Println("  • Reduces implementation complexity")
	fmt.Println("  • Better request tracking")
	fmt.Println("  • Clearer error semantics")
	fmt.Println()

	// Example 11: Migration Guide
	fmt.Println("🔄 Migration from 2025-03-26")
	fmt.Println("=============================")
	fmt.Println()
	fmt.Println("Required Changes:")
	fmt.Println()
	fmt.Println("1. Remove batch request code:")
	fmt.Println("   - Remove BatchCall() implementations")
	fmt.Println("   - Remove ReadBatch()/WriteBatch() methods")
	fmt.Println("   - Remove batch detection logic")
	fmt.Println()
	fmt.Println("2. Add output schemas to tools:")
	fmt.Println("   - Define OutputSchema for structured tools")
	fmt.Println("   - Use OutputSchemaFromType() for auto-generation")
	fmt.Println()
	fmt.Println("3. Add title fields:")
	fmt.Println("   - Set title on resources for display")
	fmt.Println("   - Set title on prompts for better UX")
	fmt.Println()
	fmt.Println("4. Update protocol version:")
	fmt.Println("   - Change version to \"2025-06-18\"")
	fmt.Println("   - Add MCP-Protocol-Version header for HTTP")
	fmt.Println()
	fmt.Println("5. Consider adding elicitation:")
	fmt.Println("   - Implement elicitation/create handler")
	fmt.Println("   - Add client capability declaration")
	fmt.Println()

	// Example 12: New Capabilities
	fmt.Println("🚀 New Capability Declarations")
	fmt.Println("===============================")
	fmt.Println()
	fmt.Println("Client capabilities expanded:")
	fmt.Print(`
  clientCapabilities := {
      "roots": {               // Filesystem boundaries
          "listChanged": true
      },
      "sampling": {},           // LLM requests
      "elicitation": {}         // User input requests (NEW)
  }
`)
	fmt.Println("Server capabilities unchanged:")
	fmt.Print(`
  serverCapabilities := {
      "tools": {"listChanged": true},
      "resources": {"subscribe": true, "listChanged": true},
      "prompts": {"listChanged": true},
      "completions": {}
  }
`)
	fmt.Println()

	// Example 13: Complete Example
	fmt.Println("💻 Complete Implementation Example")
	fmt.Println("===================================")
	fmt.Println()
	fmt.Println("Tool with output schema:")
	fmt.Print(`
  type SearchResult struct {
      Query   string   `+"`json:\"query\"`"+`
      Results []string `+"`json:\"results\"`"+`
      Count   int      `+"`json:\"count\"`"+`
  }

  tool := builder.NewTool("search").
      Title("Semantic Search").  // Display name
      Description("Search through documents").
      Handler(func(ctx context.Context, query string) (SearchResult, error) {
          // Implementation
          return SearchResult{
              Query:   query,
              Results: [...],
              Count:   5,
          }, nil
      }).
      OutputSchemaFromType(SearchResult{}).  // 2025-06-18
      ReadOnly(true).
      Build()
`)
	fmt.Println("Resource with metadata:")
	fmt.Print(`
  resource := &Resource{
      URI:   "file:///project/config.json",
      Name:  "project_config",
      Title: "Project Configuration",  // 2025-06-18
      Meta: map[string]interface{}{    // 2025-06-18
          "version":      "2.0",
          "lastModified": time.Now(),
          "audience":     []string{"user", "assistant"},
          "priority":     0.9,
      },
  }
`)

	// Example 14: Best Practices
	fmt.Println("📋 Best Practices for 2025-06-18")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Println("Output Schemas:")
	fmt.Println("  ✓ Define schemas for structured tool outputs")
	fmt.Println("  ✓ Use OutputSchemaFromType() for type safety")
	fmt.Println("  ✓ Include descriptions for all properties")
	fmt.Println("  ✓ Mark required fields explicitly")
	fmt.Println()
	fmt.Println("Title Fields:")
	fmt.Println("  ✓ Always provide title for user-facing items")
	fmt.Println("  ✓ Keep titles concise and descriptive")
	fmt.Println("  ✓ Use name for technical IDs, title for display")
	fmt.Println()
	fmt.Println("Metadata (_meta):")
	fmt.Println("  ✓ Use _meta for optional, extensible data")
	fmt.Println("  ✓ Include lastModified for cache management")
	fmt.Println("  ✓ Use audience for targeting")
	fmt.Println("  ✓ Use priority for ordering (0.0-1.0)")
	fmt.Println()
	fmt.Println("Elicitation:")
	fmt.Println("  ✓ Only request necessary information")
	fmt.Println("  ✓ Provide clear descriptions")
	fmt.Println("  ✓ Use appropriate JSON Schema constraints")
	fmt.Println("  ✓ Handle all three response actions")
	fmt.Println()
	fmt.Println("Version Management:")
	fmt.Println("  ✓ Always send MCP-Protocol-Version header")
	fmt.Println("  ✓ Implement version negotiation")
	fmt.Println("  ✓ Gracefully handle incompatible versions")
	fmt.Println()

	fmt.Println("✨ MCP 2025-06-18 demonstration complete!")
	fmt.Println()
	fmt.Println("Key Takeaways:")
	fmt.Println("  1. Tool output schemas enable better type safety")
	fmt.Println("  2. Title and _meta fields improve UX")
	fmt.Println("  3. Resource links connect tool results to data")
	fmt.Println("  4. Elicitation enables structured user input")
	fmt.Println("  5. Batch requests removed for simplicity")
	fmt.Println("  6. Protocol version header now required")
}
