package server

import (
	"context"
	"regexp"
	"sync"

	"github.com/jmcarbo/fullmcp/mcp"
)

// ResourceFunc reads resource content
type ResourceFunc func(context.Context) ([]byte, error)

// ResourceTemplateFunc reads resource content with parameters
type ResourceTemplateFunc func(context.Context, map[string]string) ([]byte, error)

// ResourceContent represents resource content with metadata
type ResourceContentWithMetadata struct {
	Data     []byte
	MimeType string
	URI      string
}

// ResourceHandler wraps a resource function
type ResourceHandler struct {
	URI         string
	Name        string
	Description string
	MimeType    string
	Reader      ResourceFunc
	Tags        []string
}

// ResourceTemplateHandler handles parameterized resources
type ResourceTemplateHandler struct {
	URITemplate string
	Name        string
	Description string
	MimeType    string
	Reader      ResourceTemplateFunc
	Tags        []string
	pattern     *regexp.Regexp
}

// ResourceManager manages resources
type ResourceManager struct {
	resources map[string]*ResourceHandler
	templates map[string]*ResourceTemplateHandler
	mu        sync.RWMutex
}

// NewResourceManager creates a new resource manager
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		resources: make(map[string]*ResourceHandler),
		templates: make(map[string]*ResourceTemplateHandler),
	}
}

// Register registers a resource
func (rm *ResourceManager) Register(handler *ResourceHandler) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.resources[handler.URI] = handler
	return nil
}

// RegisterTemplate registers a resource template
func (rm *ResourceManager) RegisterTemplate(handler *ResourceTemplateHandler) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Convert URI template to regex pattern
	pattern := templateToRegex(handler.URITemplate)
	compiledPattern, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	handler.pattern = compiledPattern

	rm.templates[handler.URITemplate] = handler
	return nil
}

// Read reads a resource (legacy method - returns only data)
func (rm *ResourceManager) Read(ctx context.Context, uri string) ([]byte, error) {
	content, err := rm.ReadWithMetadata(ctx, uri)
	if err != nil {
		return nil, err
	}
	return content.Data, nil
}

// ReadWithMetadata reads a resource with metadata (MIME type, etc.)
func (rm *ResourceManager) ReadWithMetadata(ctx context.Context, uri string) (*ResourceContentWithMetadata, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Try exact match first
	if handler, exists := rm.resources[uri]; exists {
		data, err := handler.Reader(ctx)
		if err != nil {
			return nil, err
		}
		mimeType := handler.MimeType
		if mimeType == "" {
			mimeType = "text/plain"
		}
		return &ResourceContentWithMetadata{
			Data:     data,
			MimeType: mimeType,
			URI:      uri,
		}, nil
	}

	// Try templates
	for _, template := range rm.templates {
		if params, ok := template.Match(uri); ok {
			data, err := template.Reader(ctx, params)
			if err != nil {
				return nil, err
			}
			mimeType := template.MimeType
			if mimeType == "" {
				mimeType = "text/plain"
			}
			return &ResourceContentWithMetadata{
				Data:     data,
				MimeType: mimeType,
				URI:      uri,
			}, nil
		}
	}

	return nil, &mcp.NotFoundError{Type: "resource", Name: uri}
}

// List returns all resources
func (rm *ResourceManager) List() []*mcp.Resource {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	resources := make([]*mcp.Resource, 0, len(rm.resources))
	for _, handler := range rm.resources {
		resources = append(resources, &mcp.Resource{
			URI:         handler.URI,
			Name:        handler.Name,
			Description: handler.Description,
			MimeType:    handler.MimeType,
		})
	}

	return resources
}

// ListTemplates returns all resource templates
func (rm *ResourceManager) ListTemplates() []*mcp.ResourceTemplate {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	templates := make([]*mcp.ResourceTemplate, 0, len(rm.templates))
	for _, handler := range rm.templates {
		templates = append(templates, &mcp.ResourceTemplate{
			URITemplate: handler.URITemplate,
			Name:        handler.Name,
			Description: handler.Description,
			MimeType:    handler.MimeType,
		})
	}

	return templates
}

// Match checks if a URI matches the template and returns parameters
func (rth *ResourceTemplateHandler) Match(uri string) (map[string]string, bool) {
	matches := rth.pattern.FindStringSubmatch(uri)
	if matches == nil {
		return nil, false
	}

	params := make(map[string]string)
	for i, name := range rth.pattern.SubexpNames() {
		if i > 0 && name != "" {
			params[name] = matches[i]
		}
	}

	return params, true
}

// templateToRegex converts a URI template to a regex pattern
// Example: "file:///{path}" -> "^file:///(?P<path>[^/]+)$"
func templateToRegex(template string) string {
	// Escape special regex characters except for {}
	escaped := regexp.QuoteMeta(template)

	// Convert {param} to named capture groups
	re := regexp.MustCompile(`\\{(\w+)\\}`)
	pattern := re.ReplaceAllString(escaped, `(?P<$1>[^/]+)`)

	// Anchor the pattern
	return "^" + pattern + "$"
}
