// Package main demonstrates tool annotations in MCP (2025-03-26 specification)
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jmcarbo/fullmcp/builder"
	"github.com/jmcarbo/fullmcp/mcp"
	"github.com/jmcarbo/fullmcp/server"
)

func main() {
	fmt.Println("MCP Tool Annotations Example")
	fmt.Println("=============================")
	fmt.Println()

	// Example 1: What are Tool Annotations?
	fmt.Println("💡 Tool Annotations Overview")
	fmt.Println("============================")
	fmt.Println()
	fmt.Println("Tool annotations are metadata hints added in MCP 2025-03-26 that help")
	fmt.Println("AI agents and clients better understand tool behavior and make smarter")
	fmt.Println("decisions about when and how to use tools.")
	fmt.Println()

	// Example 2: Available Annotations
	fmt.Println("🏷️  Available Annotations")
	fmt.Println("========================")
	fmt.Println()

	annotations := []struct {
		name        string
		description string
		useCases    string
	}{
		{
			"Title",
			"Human-readable title for the tool",
			"Display in UI, documentation",
		},
		{
			"ReadOnlyHint",
			"Tool doesn't modify environment",
			"Safe to run in parallel, no side effects",
		},
		{
			"DestructiveHint",
			"Tool may perform destructive updates",
			"Requires confirmation, careful execution",
		},
		{
			"IdempotentHint",
			"Repeated calls have no additional effect",
			"Safe to retry on failure",
		},
		{
			"OpenWorldHint",
			"Tool may interact with external entities",
			"Network calls, external APIs, unpredictable",
		},
	}

	for i, ann := range annotations {
		fmt.Printf("   %d. %s\n", i+1, ann.name)
		fmt.Printf("      %s\n", ann.description)
		fmt.Printf("      Use: %s\n", ann.useCases)
		fmt.Println()
	}

	// Example 3: Read-Only Tools
	fmt.Println("📖 Read-Only Tools")
	fmt.Println("==================")
	fmt.Println()

	readFileTool, _ := builder.NewTool("read_file").
		Title("File Reader").
		Description("Read contents of a file").
		ReadOnly(). // Mark as read-only
		Handler(func(_ context.Context, args struct {
			Path string `json:"path"`
		},
		) (string, error) {
			data, err := os.ReadFile(args.Path)
			if err != nil {
				return "", err
			}
			return string(data), nil
		}).
		Build()

	fmt.Println("Tool: read_file")
	fmt.Println("  Title:        ", readFileTool.Title)
	fmt.Println("  ReadOnlyHint: ", *readFileTool.ReadOnlyHint)
	fmt.Println()
	fmt.Println("Benefits:")
	fmt.Println("  ✓ Can be executed in parallel safely")
	fmt.Println("  ✓ No confirmation needed")
	fmt.Println("  ✓ AI can use freely without side effects")
	fmt.Println()

	// Example 4: Destructive Tools
	fmt.Println("⚠️  Destructive Tools")
	fmt.Println("====================")
	fmt.Println()

	deleteFileTool, _ := builder.NewTool("delete_file").
		Title("File Deleter").
		Description("Permanently delete a file").
		Destructive(). // Mark as destructive
		Idempotent().  // Can retry safely
		Handler(func(_ context.Context, args struct {
			Path string `json:"path"`
		},
		) (string, error) {
			if err := os.Remove(args.Path); err != nil {
				return "", err
			}
			return "File deleted successfully", nil
		}).
		Build()

	fmt.Println("Tool: delete_file")
	fmt.Println("  Title:           ", deleteFileTool.Title)
	fmt.Println("  DestructiveHint: ", *deleteFileTool.DestructiveHint)
	fmt.Println("  IdempotentHint:  ", *deleteFileTool.IdempotentHint)
	fmt.Println()
	fmt.Println("Implications:")
	fmt.Println("  ⚠  Requires user confirmation before execution")
	fmt.Println("  ⚠  AI should warn user about consequences")
	fmt.Println("  ✓  Safe to retry (idempotent)")
	fmt.Println("  ⚠  Cannot be undone")
	fmt.Println()

	// Example 5: Idempotent Tools
	fmt.Println("🔄 Idempotent Tools")
	fmt.Println("===================")
	fmt.Println()

	createDirTool, _ := builder.NewTool("create_directory").
		Title("Directory Creator").
		Description("Create a directory").
		Idempotent(). // Safe to retry
		Handler(func(_ context.Context, args struct {
			Path string `json:"path"`
		},
		) (string, error) {
			if err := os.MkdirAll(args.Path, 0o755); err != nil {
				return "", err
			}
			return "Directory created", nil
		}).
		Build()

	fmt.Println("Tool: create_directory")
	fmt.Println("  Title:          ", createDirTool.Title)
	fmt.Println("  IdempotentHint: ", *createDirTool.IdempotentHint)
	fmt.Println()
	fmt.Println("Benefits:")
	fmt.Println("  ✓ Can retry on network/timeout errors")
	fmt.Println("  ✓ No duplicate effects from multiple calls")
	fmt.Println("  ✓ Safer error recovery")
	fmt.Println()

	// Example 6: Open World Tools
	fmt.Println("🌐 Open World Tools")
	fmt.Println("===================")
	fmt.Println()

	fetchAPITool, _ := builder.NewTool("fetch_weather").
		Title("Weather Fetcher").
		Description("Fetch weather data from external API").
		OpenWorld(). // Interacts with external systems
		ReadOnly().  // Doesn't modify local state
		Handler(func(_ context.Context, args struct {
			City string `json:"city"`
		},
		) (string, error) {
			// Simulate API call
			return fmt.Sprintf("Weather for %s: Sunny, 72°F", args.City), nil
		}).
		Build()

	fmt.Println("Tool: fetch_weather")
	fmt.Println("  Title:         ", fetchAPITool.Title)
	fmt.Println("  OpenWorldHint: ", *fetchAPITool.OpenWorldHint)
	fmt.Println("  ReadOnlyHint:  ", *fetchAPITool.ReadOnlyHint)
	fmt.Println()
	fmt.Println("Characteristics:")
	fmt.Println("  ⚠  May fail due to network issues")
	fmt.Println("  ⚠  Results may change between calls")
	fmt.Println("  ⚠  Unpredictable latency")
	fmt.Println("  ✓  Doesn't modify local state")
	fmt.Println()

	// Example 7: Complex Tool
	fmt.Println("🛠️  Complex Tool Example")
	fmt.Println("========================")
	fmt.Println()

	deployTool, _ := builder.NewTool("deploy_app").
		Title("Application Deployer").
		Description("Deploy application to production").
		Destructive(). // Production changes
		OpenWorld().   // Network operations
		Handler(func(_ context.Context, args struct {
			Version string `json:"version"`
		},
		) (string, error) {
			// Simulate deployment
			return fmt.Sprintf("Deployed version %s", args.Version), nil
		}).
		Build()

	fmt.Println("Tool: deploy_app")
	fmt.Println("  Title:           ", deployTool.Title)
	fmt.Println("  DestructiveHint: ", *deployTool.DestructiveHint)
	fmt.Println("  OpenWorldHint:   ", *deployTool.OpenWorldHint)
	fmt.Println()
	fmt.Println("This tool is BOTH destructive AND open-world:")
	fmt.Println("  ⚠  Requires explicit confirmation")
	fmt.Println("  ⚠  May fail due to network/service issues")
	fmt.Println("  ⚠  Cannot easily rollback")
	fmt.Println("  ⚠  Should be used with extreme caution")
	fmt.Println()

	// Example 8: Server Registration
	fmt.Println("🖥️  Server Registration")
	fmt.Println("=======================")
	fmt.Println()

	srv := server.New("annotation-demo")
	_ = srv.AddTool(readFileTool)
	_ = srv.AddTool(deleteFileTool)
	_ = srv.AddTool(createDirTool)
	_ = srv.AddTool(fetchAPITool)
	_ = srv.AddTool(deployTool)

	fmt.Println("✓ Registered 5 tools with annotations")
	fmt.Println()
	fmt.Println("When client calls tools/list, annotations are included:")
	fmt.Println()

	ctx := context.Background()
	tools := srv.HandleMessage(ctx, &mcp.Message{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	})

	fmt.Println("Sample tool from tools/list:")
	fmt.Println(`  {
    "name": "delete_file",
    "title": "File Deleter",
    "description": "Permanently delete a file",
    "inputSchema": {...},
    "destructiveHint": true,
    "idempotentHint": true
  }`)
	fmt.Println()

	_ = tools // Use the variable

	// Example 9: AI Agent Decision Making
	fmt.Println("🤖 AI Agent Decision Making")
	fmt.Println("============================")
	fmt.Println()
	fmt.Println("How AI agents use annotations:")
	fmt.Println()

	decisions := []struct {
		scenario   string
		annotation string
		decision   string
	}{
		{
			"User asks: 'What's in config.json?'",
			"ReadOnlyHint: true",
			"✓ Execute read_file immediately (safe)",
		},
		{
			"User asks: 'Delete old logs'",
			"DestructiveHint: true",
			"⚠ Ask for confirmation first",
		},
		{
			"Network timeout during API call",
			"IdempotentHint: true",
			"✓ Retry automatically",
		},
		{
			"User asks: 'What's the weather?'",
			"OpenWorldHint: true",
			"ℹ Warn: Results may vary, network required",
		},
		{
			"Parallel data collection",
			"ReadOnlyHint: true",
			"✓ Execute all reads in parallel",
		},
	}

	for i, d := range decisions {
		fmt.Printf("   %d. Scenario: %s\n", i+1, d.scenario)
		fmt.Printf("      Annotation: %s\n", d.annotation)
		fmt.Printf("      Decision: %s\n", d.decision)
		fmt.Println()
	}

	// Example 10: Best Practices
	fmt.Println("📋 Best Practices")
	fmt.Println("=================")
	fmt.Println()
	fmt.Println("When to use each annotation:")
	fmt.Println()
	fmt.Println("ReadOnlyHint:")
	fmt.Println("  • GET operations, queries, reads")
	fmt.Println("  • Calculations, validations, checks")
	fmt.Println("  • Any tool with NO side effects")
	fmt.Println()
	fmt.Println("DestructiveHint:")
	fmt.Println("  • DELETE operations")
	fmt.Println("  • Irreversible changes")
	fmt.Println("  • Production deployments")
	fmt.Println("  • Data modifications")
	fmt.Println()
	fmt.Println("IdempotentHint:")
	fmt.Println("  • PUT/PATCH with same result")
	fmt.Println("  • Create if not exists")
	fmt.Println("  • Safe to retry operations")
	fmt.Println()
	fmt.Println("OpenWorldHint:")
	fmt.Println("  • External API calls")
	fmt.Println("  • Network operations")
	fmt.Println("  • Third-party service interactions")
	fmt.Println("  • Non-deterministic results")
	fmt.Println()

	// Example 11: Annotation Combinations
	fmt.Println("🔀 Common Annotation Combinations")
	fmt.Println("==================================")
	fmt.Println()

	combinations := []struct {
		combo       string
		description string
		example     string
	}{
		{
			"ReadOnly + OpenWorld",
			"Safe queries to external services",
			"fetch_user_info, get_stock_price",
		},
		{
			"Destructive + Idempotent",
			"Destructive but safe to retry",
			"delete_file, drop_table",
		},
		{
			"Destructive + OpenWorld",
			"External destructive operations",
			"deploy_app, charge_credit_card",
		},
		{
			"ReadOnly only",
			"Pure local reads",
			"read_file, list_directory",
		},
	}

	for i, c := range combinations {
		fmt.Printf("   %d. %s\n", i+1, c.combo)
		fmt.Printf("      %s\n", c.description)
		fmt.Printf("      Examples: %s\n", c.example)
		fmt.Println()
	}

	// Example 12: UI/UX Implications
	fmt.Println("🎨 UI/UX Implications")
	fmt.Println("=====================")
	fmt.Println()
	fmt.Println("How clients use annotations for better UX:")
	fmt.Println()
	fmt.Println("ReadOnlyHint:")
	fmt.Println("  → Execute immediately without prompts")
	fmt.Println("  → Show in 'Safe Tools' category")
	fmt.Println("  → Green icon/indicator")
	fmt.Println()
	fmt.Println("DestructiveHint:")
	fmt.Println("  → Show confirmation dialog")
	fmt.Println("  → Require explicit user approval")
	fmt.Println("  → Red icon/warning indicator")
	fmt.Println("  → Log for audit trail")
	fmt.Println()
	fmt.Println("IdempotentHint:")
	fmt.Println("  → Auto-retry on transient failures")
	fmt.Println("  → Show 'Safe to retry' in docs")
	fmt.Println()
	fmt.Println("OpenWorldHint:")
	fmt.Println("  → Show network indicator")
	fmt.Println("  → Warn if offline")
	fmt.Println("  → Add timeout indicators")
	fmt.Println()

	fmt.Println("✨ Tool annotations demonstration complete!")
	fmt.Println()
	fmt.Println("Key Takeaways:")
	fmt.Println("  1. Annotations help AI make smarter decisions")
	fmt.Println("  2. Use ReadOnlyHint for safe parallel execution")
	fmt.Println("  3. Use DestructiveHint to require confirmation")
	fmt.Println("  4. Use IdempotentHint for retry-safe operations")
	fmt.Println("  5. Use OpenWorldHint for external dependencies")
	fmt.Println("  6. Combine annotations for precise semantics")
}
