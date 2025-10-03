// Package main demonstrates sampling (LLM completion) support in MCP
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jmcarbo/fullmcp/mcp"
	"github.com/jmcarbo/fullmcp/server"
)

// MockLLM simulates an LLM for demonstration purposes
func MockLLM(_ context.Context, req *mcp.CreateMessageRequest) (*mcp.CreateMessageResult, error) {
	// In a real implementation, this would call an actual LLM API
	// For demo purposes, we'll just return a canned response

	var responseText string
	if len(req.Messages) > 0 {
		lastMsg := req.Messages[len(req.Messages)-1]
		responseText = fmt.Sprintf("Mock response to: %s", lastMsg.Content.Text)
	} else {
		responseText = "Hello! I'm a mock LLM."
	}

	return &mcp.CreateMessageResult{
		Role: "assistant",
		Content: mcp.SamplingContent{
			Type: "text",
			Text: responseText,
		},
		Model:      "mock-model-v1",
		StopReason: mcp.StopReasonEndTurn,
	}, nil
}

func main() {
	fmt.Println("MCP Sampling Example")
	fmt.Println("====================")
	fmt.Println()

	// Create a server with sampling enabled
	_ = server.New("sampling-demo", server.EnableSampling())

	fmt.Println("✓ Server created with sampling capability")

	// Create a client with sampling handler
	// In a real scenario, the client would be connected to the server via a transport
	// For this demo, we're just showing the API

	// Create a sampling request using the builder API
	samplingReq := server.NewSamplingRequest().
		WithSystemPrompt("You are a helpful assistant").
		WithMaxTokens(100).
		WithTemperature(0.7).
		WithModelPreferences(
			server.NewModelPreferences("claude-3-sonnet", "gpt-4").
				WithIntelligencePriority(0.8).
				WithSpeedPriority(0.5),
		).
		AddUserMessage("What is the capital of France?")

	fmt.Println("\n📝 Created sampling request:")
	fmt.Printf("   System Prompt: %s\n", samplingReq.SystemPrompt)
	fmt.Printf("   Max Tokens: %d\n", *samplingReq.MaxTokens)
	fmt.Printf("   Temperature: %.1f\n", *samplingReq.Temperature)
	fmt.Printf("   Messages: %d\n", len(samplingReq.Messages))
	if len(samplingReq.Messages) > 0 {
		fmt.Printf("   User Message: %s\n", samplingReq.Messages[0].Content.Text)
	}

	// Simulate LLM call
	fmt.Println("\n🤖 Simulating LLM call...")
	result, err := MockLLM(context.Background(), samplingReq)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n✅ Response received:")
	fmt.Printf("   Model: %s\n", result.Model)
	fmt.Printf("   Stop Reason: %s\n", result.StopReason)
	fmt.Printf("   Response: %s\n", result.Content.Text)

	// Demonstrate multi-turn conversation
	fmt.Println("\n\n💬 Multi-turn Conversation Example")
	fmt.Println("==================================")

	conversation := server.NewSamplingRequest().
		WithSystemPrompt("You are a math tutor").
		WithMaxTokens(150).
		AddUserMessage("What is 5 + 3?").
		AddAssistantMessage("5 + 3 = 8").
		AddUserMessage("Now multiply that by 2")

	fmt.Println("\n📝 Conversation history:")
	for i, msg := range conversation.Messages {
		fmt.Printf("   %d. [%s] %s\n", i+1, msg.Role, msg.Content.Text)
	}

	result2, err := MockLLM(context.Background(), conversation)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n✅ Response:")
	fmt.Printf("   %s\n", result2.Content.Text)

	// Show model preferences
	fmt.Println("\n\n🎯 Model Preferences Example")
	fmt.Println("============================")

	prefs := server.NewModelPreferences("claude-3-opus", "gpt-4-turbo").
		WithIntelligencePriority(0.9). // Prefer more intelligent models
		WithSpeedPriority(0.3)         // Speed is less important

	req := server.NewSamplingRequest().
		WithModelPreferences(prefs).
		AddUserMessage("Explain quantum computing")

	fmt.Println("\n📝 Model Preferences:")
	fmt.Printf("   Preferred Models: %v\n", []string{"claude-3-opus", "gpt-4-turbo"})
	fmt.Printf("   Intelligence Priority: %.1f\n", *prefs.IntelligencePriority)
	fmt.Printf("   Speed Priority: %.1f\n", *prefs.SpeedPriority)

	result3, err := MockLLM(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n✅ Response:")
	fmt.Printf("   Model Used: %s\n", result3.Model)

	fmt.Println("\n\n✨ Sampling demonstration complete!")
	fmt.Println("\nNote: In a production environment, you would:")
	fmt.Println("  1. Connect client and server via a transport")
	fmt.Println("  2. Implement actual LLM API calls in the client")
	fmt.Println("  3. Use client.WithSamplingHandler() to handle server requests")
	fmt.Println("  4. Server tools can then call srv.CreateMessage() for LLM capabilities")
}
