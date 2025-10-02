// Package main demonstrates cancellation support in MCP
package main

import (
	"fmt"
	"time"

	"github.com/jmcarbo/fullmcp/mcp"
	"github.com/jmcarbo/fullmcp/server"
)

func main() {
	fmt.Println("MCP Cancellation Example")
	fmt.Println("========================")
	fmt.Println()

	// Example 1: Server with cancellation support
	fmt.Println("🚫 Server with Cancellation Support")
	fmt.Println("====================================")
	fmt.Println()

	srv := server.New("cancellation-demo", server.WithCancellation())
	fmt.Println("✓ Server created with cancellation support")
	fmt.Println()

	// Example 2: Cancellation notification structure
	fmt.Println("📝 Cancellation Notification Structure")
	fmt.Println("=======================================")
	fmt.Println()

	notification := mcp.CancelledNotification{
		RequestID: "request-123",
		Reason:    "User requested cancellation",
	}

	fmt.Printf("   RequestID: %v\n", notification.RequestID)
	fmt.Printf("   Reason:    %s\n", notification.Reason)
	fmt.Println()

	_ = srv

	// Example 3: Cancellable context pattern
	fmt.Println("⏹️  Cancellable Context Pattern")
	fmt.Println("===============================")
	fmt.Println()

	fmt.Println("Server-side implementation:")
	fmt.Print(`
  // Create cancellable context
  ctx, cancel := context.WithCancel(parentCtx)

  // Register for cancellation
  srv.RegisterCancellable(requestID, cancel)
  defer srv.UnregisterCancellable(requestID)

  // Long-running operation
  select {
  case <-ctx.Done():
    return nil, ctx.Err() // Operation was cancelled
  case result := <-workDone:
    return result, nil
  }
`)

	// Example 4: Simulating cancellation flow
	fmt.Println("🔄 Cancellation Flow Simulation")
	fmt.Println("================================")
	fmt.Println()

	fmt.Println("1. Client sends long-running request (ID: 42)")
	time.Sleep(200 * time.Millisecond)

	fmt.Println("2. Server starts processing...")
	time.Sleep(200 * time.Millisecond)

	fmt.Println("3. User clicks cancel button")
	time.Sleep(200 * time.Millisecond)

	fmt.Println("4. Client sends notifications/cancelled")
	fmt.Println("   {\"requestId\": 42, \"reason\": \"User cancelled\"}")
	time.Sleep(200 * time.Millisecond)

	fmt.Println("5. Server receives cancellation")
	time.Sleep(200 * time.Millisecond)

	fmt.Println("6. Server cancels context, stops processing")
	time.Sleep(200 * time.Millisecond)

	fmt.Println("7. Server returns error or partial result")
	fmt.Println()

	// Example 5: Race conditions
	fmt.Println("⚡ Race Condition Handling")
	fmt.Println("=========================")
	fmt.Println()

	fmt.Println("Scenario: Request completes before cancellation arrives")
	fmt.Println()

	fmt.Println("Timeline:")
	fmt.Println("   T0: Client sends request (ID: 99)")
	fmt.Println("   T1: Server processes request")
	fmt.Println("   T2: Server sends response")
	fmt.Println("   T3: User clicks cancel (too late!)")
	fmt.Println("   T4: Cancellation notification arrives")
	fmt.Println()

	fmt.Println("Handling:")
	fmt.Println("   ✓ Server: Ignore cancellation for completed request")
	fmt.Println("   ✓ Client: Ignore response after sending cancellation")
	fmt.Println()

	// Example 6: Client-side cancellation
	fmt.Println("🔌 Client-Side Cancellation")
	fmt.Println("===========================")
	fmt.Println()

	fmt.Println("Client cancels request:")
	//nolint:govet // Example code contains format directives
	fmt.Print(`
  // Send cancellation
  err := client.CancelRequest(requestID, "Operation timed out")
  if err != nil {
    log.Printf("Failed to send cancellation: %%v", err)
  }

  // Ignore any response that arrives after cancellation
`)
	fmt.Println()

	// Example 7: Server-side cancellation handling
	fmt.Println("🖥️  Server-Side Cancellation Handling")
	fmt.Println("=====================================")
	fmt.Println()

	fmt.Println("Handler with cancellation support:")
	fmt.Print(`
  func longRunningTool(ctx context.Context, args Args) (Result, error) {
    // Create cancellable context
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    // Register for cancellation
    reqID := getRequestID(ctx)
    srv.RegisterCancellable(reqID, cancel)
    defer srv.UnregisterCancellable(reqID)

    // Do work with cancellation checks
    for i := 0; i < 1000; i++ {
      select {
      case <-ctx.Done():
        return nil, ctx.Err()
      default:
        // Process item i
      }
    }

    return result, nil
  }
`)

	// Example 8: Requirements and constraints
	fmt.Println("✅ Requirements & Constraints")
	fmt.Println("=============================")
	fmt.Println()

	fmt.Println("Requirements:")
	fmt.Println("   1. MUST only cancel requests issued in same direction")
	fmt.Println("   2. SHOULD ignore responses after sending cancellation")
	fmt.Println("   3. MAY include optional reason string")
	fmt.Println()

	fmt.Println("Constraints:")
	fmt.Println("   ⚠️  Network latency may cause race conditions")
	fmt.Println("   ⚠️  Cancellation may arrive after completion")
	fmt.Println("   ⚠️  No guarantee cancellation will be processed")
	fmt.Println()

	// Example 9: Use cases
	fmt.Println("💼 Common Use Cases")
	fmt.Println("===================")
	fmt.Println()

	useCases := []struct {
		title       string
		description string
		reason      string
	}{
		{
			"User Cancellation",
			"User clicks cancel button during long operation",
			"User requested cancellation",
		},
		{
			"Timeout",
			"Operation exceeds maximum allowed time",
			"Operation timed out",
		},
		{
			"Resource Limits",
			"Server resource constraints require stopping",
			"Resource limit exceeded",
		},
		{
			"Client Shutdown",
			"Client application is closing",
			"Client shutting down",
		},
		{
			"Priority Changes",
			"Higher priority request needs resources",
			"Superseded by priority request",
		},
		{
			"Duplicate Requests",
			"User accidentally sent duplicate request",
			"Duplicate request detected",
		},
	}

	for i, uc := range useCases {
		fmt.Printf("   %d. %s\n", i+1, uc.title)
		fmt.Printf("      %s\n", uc.description)
		fmt.Printf("      Reason: \"%s\"\n", uc.reason)
		fmt.Println()
	}

	// Example 10: Best practices
	fmt.Println("📋 Best Practices")
	fmt.Println("=================")
	fmt.Println()

	fmt.Println("For Servers:")
	fmt.Println("   ✓ Use context.WithCancel for long-running operations")
	fmt.Println("   ✓ Register cancellation handlers early")
	fmt.Println("   ✓ Clean up resources on cancellation")
	fmt.Println("   ✓ Return appropriate error on cancellation")
	fmt.Println("   ✓ Check ctx.Done() periodically in loops")
	fmt.Println()

	fmt.Println("For Clients:")
	fmt.Println("   ✓ Provide clear cancellation reasons")
	fmt.Println("   ✓ Handle race conditions gracefully")
	fmt.Println("   ✓ Don't rely on cancellation being processed")
	fmt.Println("   ✓ Implement UI for user-initiated cancellations")
	fmt.Println()

	// Example 11: Protocol flow
	fmt.Println("🔄 Protocol Flow")
	fmt.Println("================")
	fmt.Println()

	fmt.Println("1. Client sends request with ID")
	fmt.Println("2. Server starts processing with cancellable context")
	fmt.Println("3. Client sends notifications/cancelled")
	fmt.Println("4. Server receives notification, cancels context")
	fmt.Println("5. Server stops processing, cleans up resources")
	fmt.Println("6. Server may return error or ignore (best effort)")
	fmt.Println()

	fmt.Println("✨ Cancellation demonstration complete!")
	fmt.Println()
	fmt.Println("Note: In a production environment:")
	fmt.Println("  1. Enable cancellation with WithCancellation()")
	fmt.Println("  2. Use cancellable contexts in long-running handlers")
	fmt.Println("  3. Register cancel functions with RegisterCancellable()")
	fmt.Println("  4. Send notifications/cancelled from client when needed")
	fmt.Println("  5. Handle race conditions and partial results gracefully")
}
