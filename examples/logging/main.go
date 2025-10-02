// Package main demonstrates logging protocol extensions in MCP
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jmcarbo/fullmcp/mcp"
	"github.com/jmcarbo/fullmcp/server"
)

func main() {
	fmt.Println("MCP Logging Example")
	fmt.Println("===================")
	fmt.Println()

	// Example 1: Server with logging capability
	fmt.Println("📋 Server with Logging Capability")
	fmt.Println("==================================")
	fmt.Println()

	srv := server.New("logging-demo", server.EnableLogging())

	fmt.Println("✓ Server created with logging capability")
	fmt.Println()

	// Example 2: Log levels
	fmt.Println("📊 Log Levels (RFC 5424)")
	fmt.Println("========================")
	fmt.Println()

	levels := []mcp.LogLevel{
		mcp.LogLevelDebug,
		mcp.LogLevelInfo,
		mcp.LogLevelNotice,
		mcp.LogLevelWarning,
		mcp.LogLevelError,
		mcp.LogLevelCritical,
		mcp.LogLevelAlert,
		mcp.LogLevelEmergency,
	}

	for i, level := range levels {
		fmt.Printf("   %d. %s (value: %d)\n", i+1, level, level.Value())
	}
	fmt.Println()

	// Example 3: Setting minimum log level
	fmt.Println("⚙️  Setting Minimum Log Level")
	fmt.Println("=============================")
	fmt.Println()

	ctx := context.Background()

	// Client would send this request
	fmt.Println("Client sends: logging/setLevel with level=\"info\"")
	err := srv.SetLogLevel(ctx, mcp.LogLevelInfo)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Minimum log level set to INFO")
	fmt.Println()

	// Example 4: Log message filtering
	fmt.Println("🔍 Log Message Filtering")
	fmt.Println("========================")
	fmt.Println()

	minLevel := mcp.LogLevelInfo
	testMessages := []struct {
		level   mcp.LogLevel
		message string
	}{
		{mcp.LogLevelDebug, "Debug message"},
		{mcp.LogLevelInfo, "Info message"},
		{mcp.LogLevelWarning, "Warning message"},
		{mcp.LogLevelError, "Error message"},
	}

	fmt.Printf("Minimum level: %s\n", minLevel)
	fmt.Println()
	for _, msg := range testMessages {
		shouldLog := msg.level.ShouldLog(minLevel)
		status := "❌ FILTERED"
		if shouldLog {
			status = "✓ LOGGED"
		}
		fmt.Printf("   %s: %s - %s\n", status, msg.level, msg.message)
	}
	fmt.Println()

	// Example 5: Structured log messages
	fmt.Println("📝 Structured Log Messages")
	fmt.Println("==========================")
	fmt.Println()

	// Simulate logging different events
	logMessages := []struct {
		level  mcp.LogLevel
		logger string
		data   map[string]interface{}
	}{
		{
			level:  mcp.LogLevelInfo,
			logger: "server",
			data: map[string]interface{}{
				"event":   "startup",
				"version": "1.0.0",
				"port":    8080,
			},
		},
		{
			level:  mcp.LogLevelWarning,
			logger: "database",
			data: map[string]interface{}{
				"event":       "slow_query",
				"duration_ms": 1500,
				"query":       "SELECT * FROM users",
			},
		},
		{
			level:  mcp.LogLevelError,
			logger: "api",
			data: map[string]interface{}{
				"event":       "request_failed",
				"status_code": 500,
				"error":       "Internal server error",
				"path":        "/api/users",
			},
		},
		{
			level:  mcp.LogLevelCritical,
			logger: "system",
			data: map[string]interface{}{
				"event":          "disk_full",
				"available_mb":   10,
				"threshold_mb":   100,
				"filesystem":     "/var/log",
				"action_required": true,
			},
		},
	}

	for _, msg := range logMessages {
		fmt.Printf("   [%s] %s:\n", msg.level, msg.logger)
		for key, value := range msg.data {
			fmt.Printf("      %s: %v\n", key, value)
		}
		fmt.Println()
	}

	// Example 6: Client-side log handler
	fmt.Println("🔌 Client-Side Log Handler")
	fmt.Println("==========================")
	fmt.Println()

	fmt.Println("Client configuration example:")
	//nolint:govet // Example code contains format directives
	fmt.Print(`
  client := client.New(transport,
    client.WithLogHandler(func(ctx context.Context, msg *mcp.LogMessage) {
      log.Printf("[%%s] %%s: %%v", msg.Level, msg.Logger, msg.Data)
    }),
  )

  // Set minimum log level
  client.SetLogLevel(ctx, mcp.LogLevelInfo)
`)
	fmt.Println()
	fmt.Println()

	// Example 7: Server logging API
	fmt.Println("🖥️  Server Logging API")
	fmt.Println("======================")
	fmt.Println()

	fmt.Println("Server-side logging methods:")
	fmt.Print(`
  // Generic logging
  srv.Log(mcp.LogLevelInfo, "mylogger", map[string]interface{}{
    "message": "Something happened",
  })

  // Convenience methods
  srv.LogDebug("debug-logger", data)
  srv.LogInfo("info-logger", data)
  srv.LogWarning("warn-logger", data)
  srv.LogError("error-logger", data)
`)
	fmt.Println()

	// Example 8: Real-world use cases
	fmt.Println("💼 Real-World Use Cases")
	fmt.Println("=======================")
	fmt.Println()

	useCases := []struct {
		title       string
		description string
	}{
		{
			"Development/Debugging",
			"Set level to DEBUG to see detailed execution flow",
		},
		{
			"Production Monitoring",
			"Set level to WARNING to track issues without noise",
		},
		{
			"Performance Analysis",
			"Log slow queries, high memory usage, etc.",
		},
		{
			"Audit Trail",
			"Track important events (user actions, system changes)",
		},
		{
			"Error Tracking",
			"Capture errors with full context for debugging",
		},
		{
			"Security Events",
			"Log authentication failures, suspicious activity",
		},
	}

	for i, uc := range useCases {
		fmt.Printf("   %d. %s\n", i+1, uc.title)
		fmt.Printf("      %s\n", uc.description)
		fmt.Println()
	}

	// Example 9: Protocol flow
	fmt.Println("🔄 Protocol Flow")
	fmt.Println("================")
	fmt.Println()

	fmt.Println("1. Server declares logging capability during initialize")
	fmt.Println("2. Client sends logging/setLevel request")
	fmt.Println("3. Server starts emitting log notifications")
	fmt.Println("4. Server sends notifications/message for each log")
	fmt.Println("5. Client handles notifications with registered handler")
	fmt.Println()

	fmt.Println("✨ Logging demonstration complete!")
	fmt.Println()
	fmt.Println("Note: In a production environment:")
	fmt.Println("  1. Enable logging capability on server with EnableLogging()")
	fmt.Println("  2. Client sets desired log level with SetLogLevel()")
	fmt.Println("  3. Server logs events using srv.Log() or convenience methods")
	fmt.Println("  4. Client receives and processes log notifications")
	fmt.Println("  5. Use structured data for easy parsing and analysis")
}
