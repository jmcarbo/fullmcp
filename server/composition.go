package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmcarbo/fullmcp/mcp"
)

// CompositeServer allows mounting multiple servers under different namespaces
type CompositeServer struct {
	*Server
	mounted map[string]*Server
}

// NewCompositeServer creates a new composite server
func NewCompositeServer(name string, opts ...Option) *CompositeServer {
	return &CompositeServer{
		Server:  New(name, opts...),
		mounted: make(map[string]*Server),
	}
}

// Mount mounts a server at the given prefix
// All tools, resources, and prompts from the mounted server will be accessible
// with the prefix prepended to their names/URIs
func (cs *CompositeServer) Mount(prefix string, server *Server) error {
	// Normalize prefix
	prefix = strings.TrimPrefix(prefix, "/")
	prefix = strings.TrimSuffix(prefix, "/")

	if prefix == "" {
		return fmt.Errorf("mount prefix cannot be empty")
	}

	if _, exists := cs.mounted[prefix]; exists {
		return fmt.Errorf("server already mounted at prefix: %s", prefix)
	}

	cs.mounted[prefix] = server
	return nil
}

// Unmount removes a mounted server
func (cs *CompositeServer) Unmount(prefix string) error {
	prefix = strings.TrimPrefix(prefix, "/")
	prefix = strings.TrimSuffix(prefix, "/")

	if _, exists := cs.mounted[prefix]; !exists {
		return fmt.Errorf("no server mounted at prefix: %s", prefix)
	}

	delete(cs.mounted, prefix)
	return nil
}

// ListTools returns all tools from this server and mounted servers
func (cs *CompositeServer) ListTools(ctx context.Context) []*mcp.Tool {
	// Get local tools
	tools, _ := cs.Server.tools.List(ctx)

	// Get tools from mounted servers
	for prefix, server := range cs.mounted {
		mountedTools, _ := server.tools.List(ctx)

		// Prefix tool names
		for _, tool := range mountedTools {
			prefixedTool := *tool
			prefixedTool.Name = fmt.Sprintf("%s/%s", prefix, tool.Name)
			tools = append(tools, &prefixedTool)
		}
	}

	return tools
}

// CallTool executes a tool from this server or a mounted server
func (cs *CompositeServer) CallTool(ctx context.Context, name string, arguments json.RawMessage) (interface{}, error) {
	// Check if it's a prefixed tool from a mounted server
	for prefix, server := range cs.mounted {
		if strings.HasPrefix(name, prefix+"/") {
			// Strip prefix and call mounted server's tool
			toolName := strings.TrimPrefix(name, prefix+"/")
			return server.tools.Call(ctx, toolName, arguments)
		}
	}

	// Call local tool
	return cs.Server.tools.Call(ctx, name, arguments)
}

// ListResources returns all resources from this server and mounted servers
func (cs *CompositeServer) ListResources() []*mcp.Resource {
	// Get local resources
	resources := cs.Server.resources.List()

	// Get resources from mounted servers
	for prefix, server := range cs.mounted {
		mountedResources := server.resources.List()

		// Prefix resource URIs
		for _, res := range mountedResources {
			prefixedRes := *res
			prefixedRes.URI = fmt.Sprintf("%s/%s", prefix, res.URI)
			resources = append(resources, &prefixedRes)
		}
	}

	return resources
}

// ReadResource reads a resource from this server or a mounted server
func (cs *CompositeServer) ReadResource(ctx context.Context, uri string) ([]byte, error) {
	// Check if it's a prefixed resource from a mounted server
	for prefix, server := range cs.mounted {
		if strings.HasPrefix(uri, prefix+"/") {
			// Strip prefix and read from mounted server
			resourceURI := strings.TrimPrefix(uri, prefix+"/")
			return server.resources.Read(ctx, resourceURI)
		}
	}

	// Read local resource
	return cs.Server.resources.Read(ctx, uri)
}

// ListResourceTemplates returns all resource templates from this server and mounted servers
func (cs *CompositeServer) ListResourceTemplates() []*mcp.ResourceTemplate {
	// Get local templates
	templates := cs.Server.resources.ListTemplates()

	// Get templates from mounted servers
	for prefix, server := range cs.mounted {
		mountedTemplates := server.resources.ListTemplates()

		// Prefix template URIs
		for _, tmpl := range mountedTemplates {
			prefixedTmpl := *tmpl
			prefixedTmpl.URITemplate = fmt.Sprintf("%s/%s", prefix, tmpl.URITemplate)
			templates = append(templates, &prefixedTmpl)
		}
	}

	return templates
}

// ListPrompts returns all prompts from this server and mounted servers
func (cs *CompositeServer) ListPrompts() []*mcp.Prompt {
	// Get local prompts
	prompts := cs.Server.prompts.List()

	// Get prompts from mounted servers
	for prefix, server := range cs.mounted {
		mountedPrompts := server.prompts.List()

		// Prefix prompt names
		for _, prompt := range mountedPrompts {
			prefixedPrompt := *prompt
			prefixedPrompt.Name = fmt.Sprintf("%s/%s", prefix, prompt.Name)
			prompts = append(prompts, &prefixedPrompt)
		}
	}

	return prompts
}

// GetPrompt gets a prompt from this server or a mounted server
func (cs *CompositeServer) GetPrompt(ctx context.Context, name string, arguments map[string]interface{}) ([]*mcp.PromptMessage, error) {
	// Check if it's a prefixed prompt from a mounted server
	for prefix, server := range cs.mounted {
		if strings.HasPrefix(name, prefix+"/") {
			// Strip prefix and get from mounted server
			promptName := strings.TrimPrefix(name, prefix+"/")
			return server.prompts.Get(ctx, promptName, arguments)
		}
	}

	// Get local prompt
	return cs.Server.prompts.Get(ctx, name, arguments)
}

// GetMountedServers returns a map of mounted servers by prefix
func (cs *CompositeServer) GetMountedServers() map[string]*Server {
	result := make(map[string]*Server, len(cs.mounted))
	for k, v := range cs.mounted {
		result[k] = v
	}
	return result
}
