// Package main demonstrates ping utility in MCP
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/jmcarbo/fullmcp/server"
)

func main() {
	fmt.Println("MCP Ping Utility Example")
	fmt.Println("=========================")
	fmt.Println()

	// Example 1: Basic ping
	fmt.Println("🏓 Basic Ping")
	fmt.Println("=============")
	fmt.Println()

	srv := server.New("ping-demo")
	fmt.Println("✓ Server created (ping is always available)")
	fmt.Println()

	_ = srv

	// Example 2: Ping purpose
	fmt.Println("🎯 Purpose of Ping")
	fmt.Println("==================")
	fmt.Println()

	purposes := []string{
		"Keep connections alive",
		"Detect dead connections",
		"Verify server responsiveness",
		"Monitor connection latency",
		"Implement connection health checks",
	}

	for i, purpose := range purposes {
		fmt.Printf("   %d. %s\n", i+1, purpose)
	}
	fmt.Println()

	// Example 3: Protocol flow
	fmt.Println("🔄 Protocol Flow")
	fmt.Println("================")
	fmt.Println()

	fmt.Println("1. Client sends ping request:")
	fmt.Println("   {\"jsonrpc\": \"2.0\", \"method\": \"ping\", \"id\": 1}")
	fmt.Println()

	time.Sleep(200 * time.Millisecond)

	fmt.Println("2. Server responds with empty result:")
	fmt.Println("   {\"jsonrpc\": \"2.0\", \"result\": {}, \"id\": 1}")
	fmt.Println()

	// Example 4: Client-side usage
	fmt.Println("🔌 Client-Side Usage")
	fmt.Println("====================")
	fmt.Println()

	fmt.Println("Simple ping:")
	var sb1 strings.Builder
	sb1.WriteString("\n  err := client.Ping(ctx)\n")
	sb1.WriteString("  if err != nil {\n")
	sb1.WriteString("    log.Printf(\"Server unreachable: %v\", err)\n")
	sb1.WriteString("  }\n")
	fmt.Print(sb1.String())
	fmt.Println()

	// Example 5: Connection health monitoring
	fmt.Println("❤️  Connection Health Monitoring")
	fmt.Println("================================")
	fmt.Println()

	fmt.Println("Periodic health check example:")
	var sb2 strings.Builder
	sb2.WriteString("\n  ticker := time.NewTicker(30 * time.Second)\n")
	sb2.WriteString("  defer ticker.Stop()\n\n")
	sb2.WriteString("  for {\n")
	sb2.WriteString("    select {\n")
	sb2.WriteString("    case <-ticker.C:\n")
	sb2.WriteString("      start := time.Now()\n")
	sb2.WriteString("      err := client.Ping(ctx)\n")
	sb2.WriteString("      latency := time.Since(start)\n\n")
	sb2.WriteString("      if err != nil {\n")
	sb2.WriteString("        log.Printf(\"Health check failed: %v\", err)\n")
	sb2.WriteString("        // Reconnect logic here\n")
	sb2.WriteString("      } else {\n")
	sb2.WriteString("        log.Printf(\"Health check OK (latency: %v)\", latency)\n")
	sb2.WriteString("      }\n\n")
	sb2.WriteString("    case <-ctx.Done():\n")
	sb2.WriteString("      return\n")
	sb2.WriteString("    }\n")
	sb2.WriteString("  }\n")
	fmt.Print(sb2.String())

	// Example 6: Timeout handling
	fmt.Println("⏱️  Timeout Handling")
	fmt.Println("====================")
	fmt.Println()

	fmt.Println("Ping with timeout:")
	var sb3 strings.Builder
	sb3.WriteString("\n  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)\n")
	sb3.WriteString("  defer cancel()\n\n")
	sb3.WriteString("  err := client.Ping(ctx)\n")
	sb3.WriteString("  if err != nil {\n")
	sb3.WriteString("    if errors.Is(err, context.DeadlineExceeded) {\n")
	sb3.WriteString("      log.Println(\"Ping timeout - server not responding\")\n")
	sb3.WriteString("    } else {\n")
	sb3.WriteString("      log.Printf(\"Ping failed: %v\", err)\n")
	sb3.WriteString("    }\n")
	sb3.WriteString("  }\n")
	fmt.Print(sb3.String())

	// Example 7: Latency measurement
	fmt.Println("📊 Latency Measurement")
	fmt.Println("======================")
	fmt.Println()

	fmt.Println("Simulating ping latency measurements...")
	fmt.Println()

	latencies := []time.Duration{
		15 * time.Millisecond,
		12 * time.Millisecond,
		18 * time.Millisecond,
		14 * time.Millisecond,
		16 * time.Millisecond,
	}

	var total time.Duration
	for i, latency := range latencies {
		fmt.Printf("   Ping %d: %v\n", i+1, latency)
		total += latency
		time.Sleep(100 * time.Millisecond)
	}

	avg := total / time.Duration(len(latencies))
	fmt.Printf("\n   Average latency: %v\n", avg)
	fmt.Println()

	// Example 8: Use cases
	fmt.Println("💼 Common Use Cases")
	fmt.Println("===================")
	fmt.Println()

	useCases := []struct {
		title       string
		description string
	}{
		{
			"Keep-Alive",
			"Prevent idle connections from timing out",
		},
		{
			"Health Monitoring",
			"Continuous monitoring of connection health",
		},
		{
			"Failover Detection",
			"Quickly detect when to switch to backup server",
		},
		{
			"Load Balancing",
			"Measure server response times for routing decisions",
		},
		{
			"Connection Validation",
			"Verify connection before critical operations",
		},
		{
			"Debugging",
			"Test connectivity during troubleshooting",
		},
	}

	for i, uc := range useCases {
		fmt.Printf("   %d. %s\n", i+1, uc.title)
		fmt.Printf("      %s\n", uc.description)
		fmt.Println()
	}

	// Example 9: Best practices
	fmt.Println("📋 Best Practices")
	fmt.Println("=================")
	fmt.Println()

	fmt.Println("For Clients:")
	fmt.Println("   ✓ Use reasonable ping intervals (e.g., 30-60 seconds)")
	fmt.Println("   ✓ Always use timeouts with ping requests")
	fmt.Println("   ✓ Track latency trends, not just single values")
	fmt.Println("   ✓ Implement exponential backoff on failures")
	fmt.Println("   ✓ Log health check results for monitoring")
	fmt.Println()

	fmt.Println("For Servers:")
	fmt.Println("   ✓ Respond to ping quickly (minimal processing)")
	fmt.Println("   ✓ Don't perform heavy operations in ping handler")
	fmt.Println("   ✓ Ping is always available (no special setup needed)")
	fmt.Println()

	// Example 10: Advanced patterns
	fmt.Println("🔬 Advanced Patterns")
	fmt.Println("====================")
	fmt.Println()

	fmt.Println("Adaptive ping interval:")
	var sb4 strings.Builder
	sb4.WriteString("\n  interval := 30 * time.Second\n")
	sb4.WriteString("  failures := 0\n\n")
	sb4.WriteString("  for {\n")
	sb4.WriteString("    err := client.Ping(ctx)\n")
	sb4.WriteString("    if err != nil {\n")
	sb4.WriteString("      failures++\n")
	sb4.WriteString("      // Increase ping frequency on failures\n")
	sb4.WriteString("      interval = max(5*time.Second, interval/2)\n")
	sb4.WriteString("    } else {\n")
	sb4.WriteString("      failures = 0\n")
	sb4.WriteString("      // Decrease ping frequency on success\n")
	sb4.WriteString("      interval = min(60*time.Second, interval*2)\n")
	sb4.WriteString("    }\n\n")
	sb4.WriteString("    time.Sleep(interval)\n")
	sb4.WriteString("  }\n")
	fmt.Print(sb4.String())

	fmt.Println("✨ Ping demonstration complete!")
	fmt.Println()
	fmt.Println("Note:")
	fmt.Println("  - Ping is always available on all MCP servers")
	fmt.Println("  - No special setup or capabilities required")
	fmt.Println("  - Returns empty object on success")
	fmt.Println("  - Either client or server can initiate ping")
	fmt.Println("  - Use for connection health monitoring")
}
