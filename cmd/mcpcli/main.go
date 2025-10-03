// Package main provides a CLI tool for MCP server management
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jmcarbo/fullmcp/client"
	"github.com/jmcarbo/fullmcp/mcp"
	"github.com/jmcarbo/fullmcp/transport/http"
	"github.com/jmcarbo/fullmcp/transport/stdio"
	"github.com/jmcarbo/fullmcp/transport/streamhttp"
	"github.com/spf13/cobra"
)

var (
	version       = "1.0.0"
	timeout       int
	verbose       bool
	url           string
	useStreamHTTP bool
	apiKey        string
)

// createTransport creates the appropriate transport based on the URL flag
func createTransport() (io.ReadWriteCloser, error) {
	if url != "" {
		if useStreamHTTP {
			// Use streamhttp transport (HTTP+SSE)
			opts := []streamhttp.Option{}
			if apiKey != "" {
				opts = append(opts, streamhttp.WithAPIKey(apiKey))
			}
			transport := streamhttp.New(url, opts...)
			return transport.Connect(context.Background())
		}
		// Use basic HTTP transport
		opts := []http.Option{}
		if apiKey != "" {
			opts = append(opts, http.WithAPIKey(apiKey))
		}
		transport := http.New(url, opts...)
		return transport.Connect(context.Background())
	}
	// Use stdio transport
	return stdio.New(), nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "mcpcli",
		Short: "MCP CLI - Model Context Protocol command-line tool",
		Long: `mcpcli is a command-line tool for interacting with MCP servers.
It supports testing connections, listing capabilities, and invoking tools.`,
		Version: version,
	}

	rootCmd.PersistentFlags().IntVarP(&timeout, "timeout", "t", 30, "Request timeout in seconds")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&url, "url", "u", "", "MCP server URL (use HTTP transport instead of stdio)")
	rootCmd.PersistentFlags().BoolVar(&useStreamHTTP, "stream", false, "Use streamhttp transport (HTTP+SSE) instead of basic HTTP")
	rootCmd.PersistentFlags().StringVarP(&apiKey, "api-key", "k", "", "API key for authentication (sent as X-API-Key header)")

	// Add commands
	rootCmd.AddCommand(pingCmd())
	rootCmd.AddCommand(listToolsCmd())
	rootCmd.AddCommand(listResourcesCmd())
	rootCmd.AddCommand(listPromptsCmd())
	rootCmd.AddCommand(callToolCmd())
	rootCmd.AddCommand(readResourceCmd())
	rootCmd.AddCommand(getPromptCmd())
	rootCmd.AddCommand(infoCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func pingCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ping",
		Short: "Test connection to an MCP server",
		Long:  `Establishes a connection to an MCP server and verifies it responds.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			transport, err := createTransport()
			if err != nil {
				return fmt.Errorf("failed to create transport: %w", err)
			}
			c := client.New(transport)

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			defer cancel()

			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer func() { _ = c.Close() }()

			fmt.Println("✓ Successfully connected to MCP server")
			return nil
		},
	}
}

func listToolsCmd() *cobra.Command {
	var outputJSON bool

	cmd := &cobra.Command{
		Use:   "list-tools",
		Short: "List available tools",
		Long:  `Retrieves and displays all tools available on the MCP server.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			transport, err := createTransport()
			if err != nil {
				return fmt.Errorf("failed to create transport: %w", err)
			}
			c := client.New(transport)

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			defer cancel()

			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer func() { _ = c.Close() }()

			tools, err := c.ListTools(ctx)
			if err != nil {
				return fmt.Errorf("failed to list tools: %w", err)
			}

			if outputJSON {
				data, _ := json.MarshalIndent(tools, "", "  ")
				fmt.Println(string(data))
			} else {
				fmt.Printf("Available Tools (%d):\n\n", len(tools))
				for _, tool := range tools {
					fmt.Printf("  • %s\n", tool.Name)
					if tool.Description != "" {
						fmt.Printf("    %s\n", tool.Description)
					}
					if verbose && tool.InputSchema != nil {
						schema, _ := json.MarshalIndent(tool.InputSchema, "    ", "  ")
						fmt.Printf("    Schema: %s\n", string(schema))
					}
					fmt.Println()
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func listResourcesCmd() *cobra.Command {
	var outputJSON bool

	cmd := &cobra.Command{
		Use:   "list-resources",
		Short: "List available resources",
		Long:  `Retrieves and displays all resources available on the MCP server.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			transport, err := createTransport()
			if err != nil {
				return fmt.Errorf("failed to create transport: %w", err)
			}
			c := client.New(transport)

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			defer cancel()

			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer func() { _ = c.Close() }()

			resources, err := c.ListResources(ctx)
			if err != nil {
				return fmt.Errorf("failed to list resources: %w", err)
			}

			if outputJSON {
				data, _ := json.MarshalIndent(resources, "", "  ")
				fmt.Println(string(data))
			} else {
				fmt.Printf("Available Resources (%d):\n\n", len(resources))
				for _, resource := range resources {
					fmt.Printf("  • %s\n", resource.URI)
					if resource.Name != "" {
						fmt.Printf("    Name: %s\n", resource.Name)
					}
					if resource.Description != "" {
						fmt.Printf("    Description: %s\n", resource.Description)
					}
					if resource.MimeType != "" {
						fmt.Printf("    MIME Type: %s\n", resource.MimeType)
					}
					fmt.Println()
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func displayPromptArguments(args []mcp.PromptArgument) {
	if !verbose || len(args) == 0 {
		return
	}

	fmt.Println("    Arguments:")
	for _, arg := range args {
		req := ""
		if arg.Required {
			req = " (required)"
		}
		fmt.Printf("      - %s%s: %s\n", arg.Name, req, arg.Description)
	}
}

func displayPromptsFormatted(prompts []*mcp.Prompt) {
	fmt.Printf("Available Prompts (%d):\n\n", len(prompts))
	for _, prompt := range prompts {
		fmt.Printf("  • %s\n", prompt.Name)
		if prompt.Description != "" {
			fmt.Printf("    %s\n", prompt.Description)
		}
		displayPromptArguments(prompt.Arguments)
		fmt.Println()
	}
}

func listPromptsCmd() *cobra.Command {
	var outputJSON bool

	cmd := &cobra.Command{
		Use:   "list-prompts",
		Short: "List available prompts",
		Long:  `Retrieves and displays all prompts available on the MCP server.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			transport, err := createTransport()
			if err != nil {
				return fmt.Errorf("failed to create transport: %w", err)
			}
			c := client.New(transport)

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			defer cancel()

			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer func() { _ = c.Close() }()

			prompts, err := c.ListPrompts(ctx)
			if err != nil {
				return fmt.Errorf("failed to list prompts: %w", err)
			}

			if outputJSON {
				data, _ := json.MarshalIndent(prompts, "", "  ")
				fmt.Println(string(data))
			} else {
				displayPromptsFormatted(prompts)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func callToolCmd() *cobra.Command {
	var argsJSON string
	var outputJSON bool

	cmd := &cobra.Command{
		Use:   "call-tool <tool-name>",
		Short: "Call a tool on the MCP server",
		Long:  `Invokes a tool with the specified arguments and displays the result.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			toolName := args[0]

			transport, err := createTransport()
			if err != nil {
				return fmt.Errorf("failed to create transport: %w", err)
			}
			c := client.New(transport)

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			defer cancel()

			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer func() { _ = c.Close() }()

			var toolArgs json.RawMessage
			if argsJSON != "" {
				toolArgs = json.RawMessage(argsJSON)
			} else {
				toolArgs = json.RawMessage("{}")
			}

			result, err := c.CallTool(ctx, toolName, toolArgs)
			if err != nil {
				return fmt.Errorf("failed to call tool: %w", err)
			}

			if outputJSON {
				data, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(data))
			} else {
				fmt.Printf("Tool Result:\n%v\n", result)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&argsJSON, "args", "", "Tool arguments as JSON")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func readResourceCmd() *cobra.Command {
	var outputJSON bool

	cmd := &cobra.Command{
		Use:   "read-resource <uri>",
		Short: "Read a resource from the MCP server",
		Long:  `Retrieves and displays the content of a resource.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			uri := args[0]

			transport, err := createTransport()
			if err != nil {
				return fmt.Errorf("failed to create transport: %w", err)
			}
			c := client.New(transport)

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			defer cancel()

			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer func() { _ = c.Close() }()

			content, err := c.ReadResource(ctx, uri)
			if err != nil {
				return fmt.Errorf("failed to read resource: %w", err)
			}

			if outputJSON {
				data, _ := json.MarshalIndent(string(content), "", "  ")
				fmt.Println(string(data))
			} else {
				fmt.Printf("Resource Content:\n%s\n", string(content))
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func getPromptCmd() *cobra.Command {
	var argsMap map[string]string
	var outputJSON bool

	cmd := &cobra.Command{
		Use:   "get-prompt <prompt-name>",
		Short: "Get a prompt from the MCP server",
		Long:  `Retrieves a prompt with the specified arguments.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			promptName := args[0]

			transport, err := createTransport()
			if err != nil {
				return fmt.Errorf("failed to create transport: %w", err)
			}
			c := client.New(transport)

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			defer cancel()

			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer func() { _ = c.Close() }()

			if argsMap == nil {
				argsMap = make(map[string]string)
			}

			// Convert to map[string]interface{}
			promptArgs := make(map[string]interface{})
			for k, v := range argsMap {
				promptArgs[k] = v
			}

			result, err := c.GetPrompt(ctx, promptName, promptArgs)
			if err != nil {
				return fmt.Errorf("failed to get prompt: %w", err)
			}

			if outputJSON {
				data, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(data))
			} else {
				fmt.Printf("Prompt Messages:\n\n")
				for i, msg := range result {
					fmt.Printf("Message %d [%s]:\n", i+1, msg.Role)
					// Print content array
					for j, content := range msg.Content {
						fmt.Printf("  Content %d:\n", j+1)
						data, _ := json.MarshalIndent(content, "    ", "  ")
						fmt.Printf("    %s\n", string(data))
					}
					fmt.Println()
				}
			}

			return nil
		},
	}

	cmd.Flags().StringToStringVar(&argsMap, "args", nil, "Prompt arguments (e.g., --args key1=value1,key2=value2)")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func infoCmd() *cobra.Command {
	var outputJSON bool

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Display server information and capabilities",
		Long:  `Connects to the MCP server and displays detailed information about its capabilities.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			transport, err := createTransport()
			if err != nil {
				return fmt.Errorf("failed to create transport: %w", err)
			}
			c := client.New(transport)

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			defer cancel()

			if err := c.Connect(ctx); err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer func() { _ = c.Close() }()

			// Get counts
			tools, _ := c.ListTools(ctx)
			resources, _ := c.ListResources(ctx)
			prompts, _ := c.ListPrompts(ctx)

			if outputJSON {
				info := map[string]interface{}{
					"tools_count":     len(tools),
					"resources_count": len(resources),
					"prompts_count":   len(prompts),
				}
				data, _ := json.MarshalIndent(info, "", "  ")
				fmt.Println(string(data))
			} else {
				fmt.Println("MCP Server Information")
				fmt.Println("======================")
				fmt.Println()
				fmt.Printf("Tools:     %d\n", len(tools))
				fmt.Printf("Resources: %d\n", len(resources))
				fmt.Printf("Prompts:   %d\n", len(prompts))
				fmt.Println()
				fmt.Println("Use --verbose for detailed listings")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}
