// Package main demonstrates progress notifications in MCP
package main

import (
	"fmt"
	"time"

	"github.com/jmcarbo/fullmcp/mcp"
	"github.com/jmcarbo/fullmcp/server"
)

func main() {
	fmt.Println("MCP Progress Notifications Example")
	fmt.Println("===================================")
	fmt.Println()

	// Example 1: Server with progress tracking
	fmt.Println("📊 Server with Progress Tracking")
	fmt.Println("=================================")
	fmt.Println()

	srv := server.New("progress-demo", server.WithProgress())
	fmt.Println("✓ Server created with progress tracking")
	fmt.Println()

	// Example 2: Progress tokens
	fmt.Println("🔑 Progress Tokens")
	fmt.Println("==================")
	fmt.Println()

	fmt.Println("Progress tokens can be strings or integers:")
	fmt.Println("   - String: \"task-123\", \"upload-abc-def\", \"batch-process-001\"")
	fmt.Println("   - Integer: 1, 42, 12345")
	fmt.Println()
	fmt.Println("Requirements:")
	fmt.Println("   ✓ Must be unique across all active requests")
	fmt.Println("   ✓ Chosen by the sender")
	fmt.Println("   ✓ Used to correlate progress updates with requests")
	fmt.Println()

	// Example 3: Progress notification structure
	fmt.Println("📝 Progress Notification Structure")
	fmt.Println("===================================")
	fmt.Println()

	total := 100.0
	notification := mcp.ProgressNotification{
		ProgressToken: "example-task",
		Progress:      50.0,
		Total:         &total,
	}

	fmt.Printf("   ProgressToken: %v\n", notification.ProgressToken)
	fmt.Printf("   Progress:      %.1f\n", notification.Progress)
	fmt.Printf("   Total:         %.1f\n", *notification.Total)
	fmt.Println()

	// Example 4: Progress updates for a long-running task
	fmt.Println("⏳ Long-Running Task Example")
	fmt.Println("============================")
	fmt.Println()

	fmt.Println("Simulating file upload with progress updates...")
	fmt.Println()

	progressToken := "upload-large-file"
	totalBytes := 1000.0

	// Simulate progress updates
	updates := []float64{0, 250, 500, 750, 1000}
	for _, current := range updates {
		percentage := (current / totalBytes) * 100
		fmt.Printf("   [%3.0f%%] Uploaded %.0f / %.0f bytes\n", percentage, current, totalBytes)

		// In real implementation, server would send notification:
		// srv.NotifyProgress(progressToken, current, &totalBytes)

		time.Sleep(200 * time.Millisecond)
	}
	fmt.Println()
	fmt.Println("✓ Upload complete!")
	fmt.Println()

	_ = progressToken
	_ = srv

	// Example 5: Unknown total (indefinite progress)
	fmt.Println("🔄 Unknown Total Example")
	fmt.Println("========================")
	fmt.Println()

	fmt.Println("For operations where total is unknown:")
	fmt.Println()

	indeterminateToken := "process-items"
	processedItems := []int{0, 5, 12, 27, 45, 63}

	for _, count := range processedItems {
		fmt.Printf("   Processed %d items...\n", count)

		// Progress increases but total is nil
		// srv.NotifyProgress(indeterminateToken, float64(count), nil)

		time.Sleep(150 * time.Millisecond)
	}
	fmt.Println()
	fmt.Println("✓ Processing complete!")
	fmt.Println()

	_ = indeterminateToken

	// Example 6: Multiple concurrent tasks
	fmt.Println("🔀 Multiple Concurrent Tasks")
	fmt.Println("============================")
	fmt.Println()

	tasks := []struct {
		token string
		name  string
	}{
		{"task-1", "Data Import"},
		{"task-2", "Image Processing"},
		{"task-3", "Report Generation"},
	}

	fmt.Println("Managing progress for multiple tasks:")
	fmt.Println()

	for _, task := range tasks {
		fmt.Printf("   [%s] %s - Started\n", task.token, task.name)
	}
	fmt.Println()

	// Simulate concurrent progress
	time.Sleep(300 * time.Millisecond)
	fmt.Println("   [task-1] Data Import - 50% complete")
	fmt.Println("   [task-2] Image Processing - 25% complete")
	fmt.Println("   [task-3] Report Generation - 75% complete")
	fmt.Println()

	time.Sleep(300 * time.Millisecond)
	fmt.Println("   [task-1] Data Import - Complete ✓")
	fmt.Println("   [task-2] Image Processing - 75% complete")
	fmt.Println("   [task-3] Report Generation - Complete ✓")
	fmt.Println()

	// Example 7: Client-side progress handler
	fmt.Println("🔌 Client-Side Progress Handler")
	fmt.Println("================================")
	fmt.Println()

	fmt.Println("Client configuration example:")
	//nolint:govet // Example code contains format directives
	fmt.Print(`
  client := client.New(transport,
    client.WithProgressHandler(
      func(ctx context.Context, notif *mcp.ProgressNotification) {
        if notif.Total != nil {
          percent := (notif.Progress / *notif.Total) * 100
          fmt.Printf("[%%v] %%.1f%%%% complete\n",
            notif.ProgressToken, percent)
        } else {
          fmt.Printf("[%%v] Processed: %%.0f\n",
            notif.ProgressToken, notif.Progress)
        }
      },
    ),
  )
`)
	fmt.Println()
	fmt.Println()

	// Example 8: Server-side progress API
	fmt.Println("🖥️  Server-Side Progress API")
	fmt.Println("============================")
	fmt.Println()

	fmt.Println("Sending progress from server:")
	fmt.Print(`
  // With total
  total := 100.0
  srv.NotifyProgress("task-id", 50.0, &total)

  // Without total (indefinite)
  srv.NotifyProgress("task-id", 42.0, nil)

  // Using progress context in handlers
  pc := server.NewProgressContext(progressToken, srv.Progress)
  pc.Update(75.0, &total)
`)
	fmt.Println()

	// Example 9: Requirements and best practices
	fmt.Println("✅ Requirements & Best Practices")
	fmt.Println("================================")
	fmt.Println()

	fmt.Println("Requirements:")
	fmt.Println("   1. Progress value MUST increase with each notification")
	fmt.Println("   2. Progress tokens MUST be unique across active requests")
	fmt.Println("   3. Progress and total MAY be floating point values")
	fmt.Println()

	fmt.Println("Best Practices:")
	fmt.Println("   ✓ Send updates at reasonable intervals (not too frequent)")
	fmt.Println("   ✓ Include total when known for percentage calculation")
	fmt.Println("   ✓ Use descriptive token names for debugging")
	fmt.Println("   ✓ Clean up tokens when operations complete")
	fmt.Println("   ✓ Handle both determinate and indeterminate progress")
	fmt.Println()

	// Example 10: Use cases
	fmt.Println("💼 Common Use Cases")
	fmt.Println("===================")
	fmt.Println()

	useCases := []struct {
		title       string
		description string
	}{
		{
			"File Uploads/Downloads",
			"Track bytes transferred with total file size",
		},
		{
			"Batch Processing",
			"Track items processed out of total items",
		},
		{
			"Data Import/Export",
			"Show records processed with total count",
		},
		{
			"Report Generation",
			"Track stages or percentage completion",
		},
		{
			"Search/Indexing",
			"Show documents processed (may not know total)",
		},
		{
			"Long Computations",
			"Track iteration progress or time elapsed",
		},
	}

	for i, uc := range useCases {
		fmt.Printf("   %d. %s\n", i+1, uc.title)
		fmt.Printf("      %s\n", uc.description)
		fmt.Println()
	}

	// Example 11: Protocol flow
	fmt.Println("🔄 Protocol Flow")
	fmt.Println("================")
	fmt.Println()

	fmt.Println("1. Client sends request with progressToken in metadata")
	fmt.Println("2. Server starts long-running operation")
	fmt.Println("3. Server sends periodic notifications/progress updates")
	fmt.Println("4. Client receives and displays progress to user")
	fmt.Println("5. Operation completes, final response sent")
	fmt.Println()

	fmt.Println("✨ Progress notifications demonstration complete!")
	fmt.Println()
	fmt.Println("Note: In a production environment:")
	fmt.Println("  1. Include progressToken in request metadata")
	fmt.Println("  2. Server sends notifications/progress periodically")
	fmt.Println("  3. Client handles notifications with registered handler")
	fmt.Println("  4. Ensure progress values always increase")
	fmt.Println("  5. Use unique tokens for each tracked operation")
}
