// Package main demonstrates roots (filesystem boundaries) support in MCP
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jmcarbo/fullmcp/mcp"
	"github.com/jmcarbo/fullmcp/server"
)

func main() {
	fmt.Println("MCP Roots Example")
	fmt.Println("=================")
	fmt.Println()

	// Example 1: Client-side roots provider
	fmt.Println("📁 Client Roots Provider Example")
	fmt.Println("==================================")
	fmt.Println()

	// Get the current working directory as a root
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	// Create a roots provider function
	rootsProvider := func(_ context.Context) ([]mcp.Root, error) {
		return []mcp.Root{
			{
				URI:  fmt.Sprintf("file://%s", cwd),
				Name: "Current Project",
			},
			{
				URI:  fmt.Sprintf("file://%s/Documents", homeDir),
				Name: "Documents",
			},
			{
				URI:  fmt.Sprintf("file://%s/Downloads", homeDir),
				Name: "Downloads",
			},
		}, nil
	}

	fmt.Println("✓ Created roots provider with 3 roots:")
	roots, _ := rootsProvider(context.Background())
	for i, root := range roots {
		fmt.Printf("   %d. [%s] %s\n", i+1, root.Name, root.URI)
	}

	// Example 2: Using roots with a client
	fmt.Println()
	fmt.Println()
	fmt.Println("🔌 Client with Roots Support")
	fmt.Println("==============================")
	fmt.Println()

	// In a real scenario, you would create a client with the roots provider
	// For demonstration purposes, we'll show the API
	fmt.Println("Creating a client with roots support:")
	fmt.Print(`
  client := client.New(transport, client.WithRoots(rootsProvider))

  // The client will:
  // 1. Declare roots capability during initialization
  // 2. Respond to roots/list requests from the server
  // 3. Send notifications when roots change
`)
	fmt.Println()

	// Example 3: Server receiving roots notifications
	fmt.Println()
	fmt.Println("🔔 Server Roots Handler Example")
	fmt.Println("=================================")
	fmt.Println()

	rootsChanged := false
	srv := server.New(
		"roots-demo",
		server.WithRootsHandler(func(_ context.Context) {
			fmt.Println("   ⚠️  Roots have changed! Refreshing...")
			rootsChanged = true
		}),
	)

	fmt.Println("✓ Server created with roots handler")
	fmt.Println("\nWhen the client sends a roots/list_changed notification:")
	fmt.Println("   - The server's handler will be called")
	fmt.Println("   - The server can then request the updated roots list")

	// Simulate notification
	fmt.Println("\n💡 Simulating roots change notification...")
	if srv != nil {
		// In a real implementation, this would come from the client
		// For demo purposes, we'll just show the state change
		rootsChanged = true
		fmt.Printf("   Status: rootsChanged = %v\n", rootsChanged)
	}

	// Example 4: Different types of roots
	fmt.Println()
	fmt.Println()
	fmt.Println("🌐 Different Root URI Types")
	fmt.Println("============================")
	fmt.Println()

	exampleRoots := []mcp.Root{
		{
			URI:  "file:///home/user/projects/myapp",
			Name: "Main Application",
		},
		{
			URI:  "file:///var/log",
			Name: "System Logs",
		},
		{
			URI:  "https://api.example.com",
			Name: "API Endpoint",
		},
		{
			URI:  "git://github.com/user/repo.git",
			Name: "Git Repository",
		},
	}

	fmt.Println("Roots can use various URI schemes:")
	for i, root := range exampleRoots {
		fmt.Printf("   %d. %s\n", i+1, root.URI)
		fmt.Printf("      Name: %s\n", root.Name)
	}

	// Example 5: Security boundaries
	fmt.Println()
	fmt.Println()
	fmt.Println("🔒 Security and Boundaries")
	fmt.Println("===========================")
	fmt.Println()

	// Example of restricted roots for a file server
	restrictedRoots := []mcp.Root{
		{
			URI:  fmt.Sprintf("file://%s", filepath.Join(homeDir, "Documents")),
			Name: "Documents (Read-Only)",
		},
		{
			URI:  fmt.Sprintf("file://%s", filepath.Join(homeDir, "Public")),
			Name: "Public Files",
		},
	}

	fmt.Println("Security example - File server with restricted access:")
	for i, root := range restrictedRoots {
		fmt.Printf("   %d. %s\n", i+1, root.Name)
		fmt.Printf("      URI: %s\n", root.URI)
		fmt.Printf("      Purpose: Server can only access files within this boundary\n")
	}

	// Example 6: Dynamic roots
	fmt.Println()
	fmt.Println()
	fmt.Println("🔄 Dynamic Roots (Change Notifications)")
	fmt.Println("=========================================")
	fmt.Println()

	fmt.Println("Example workflow for dynamic roots:")
	fmt.Println("   1. User opens a new project in their IDE")
	fmt.Println("   2. Client updates its roots list")
	fmt.Println("   3. Client sends notifications/roots/list_changed")
	fmt.Println("   4. Server receives notification and calls roots/list")
	fmt.Println("   5. Server gets updated roots and adjusts its behavior")

	// Example 7: Practical use case
	fmt.Println()
	fmt.Println()
	fmt.Println("💼 Practical Use Case")
	fmt.Println("=====================")
	fmt.Println()

	fmt.Println("IDE Integration Example:")
	fmt.Println("   Client (IDE):")
	fmt.Println("      - Exposes workspace folders as roots")
	fmt.Println("      - Updates roots when user opens/closes projects")
	fmt.Println("      - Notifies server of changes")
	fmt.Println()
	fmt.Println("   Server (Code Analysis Tool):")
	fmt.Println("      - Requests current roots on startup")
	fmt.Println("      - Scans only files within root boundaries")
	fmt.Println("      - Refreshes analysis when roots change")
	fmt.Println("      - Respects security boundaries")

	fmt.Println()
	fmt.Println()
	fmt.Println("✨ Roots demonstration complete!")
	fmt.Println()
	fmt.Println("Note: In a production environment, you would:")
	fmt.Println("  1. Implement roots provider based on your application's workspace")
	fmt.Println("  2. Use client.WithRoots() to enable roots support")
	fmt.Println("  3. Handle roots/list requests automatically via the client")
	fmt.Println("  4. Send notifications when workspace changes")
	fmt.Println("  5. Use roots to establish security boundaries for file access")
}
