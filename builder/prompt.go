// Package builder provides fluent APIs for building MCP tools, resources, and prompts.
package builder

import (
	"github.com/jmcarbo/fullmcp/mcp"
	"github.com/jmcarbo/fullmcp/server"
)

// PromptBuilder creates prompts using a fluent API
type PromptBuilder struct {
	name        string
	description string
	arguments   []mcp.PromptArgument
	renderer    server.PromptFunc
	tags        []string
}

// NewPrompt creates a new prompt builder
func NewPrompt(name string) *PromptBuilder {
	return &PromptBuilder{name: name}
}

// Description sets the prompt description
func (pb *PromptBuilder) Description(desc string) *PromptBuilder {
	pb.description = desc
	return pb
}

// Argument adds an argument to the prompt
func (pb *PromptBuilder) Argument(name, description string, required bool) *PromptBuilder {
	pb.arguments = append(pb.arguments, mcp.PromptArgument{
		Name:        name,
		Description: description,
		Required:    required,
	})
	return pb
}

// Arguments sets all prompt arguments at once
func (pb *PromptBuilder) Arguments(args ...mcp.PromptArgument) *PromptBuilder {
	pb.arguments = args
	return pb
}

// Renderer sets the prompt renderer function
func (pb *PromptBuilder) Renderer(fn server.PromptFunc) *PromptBuilder {
	pb.renderer = fn
	return pb
}

// Tags sets the prompt tags
func (pb *PromptBuilder) Tags(tags ...string) *PromptBuilder {
	pb.tags = tags
	return pb
}

// Build creates the PromptHandler
func (pb *PromptBuilder) Build() *server.PromptHandler {
	return &server.PromptHandler{
		Name:        pb.name,
		Description: pb.description,
		Arguments:   pb.arguments,
		Renderer:    pb.renderer,
		Tags:        pb.tags,
	}
}
