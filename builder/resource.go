package builder

import (
	"context"

	"github.com/jmcarbo/fullmcp/server"
)

// ResourceBuilder creates resources using a fluent API
type ResourceBuilder struct {
	uri         string
	name        string
	description string
	mimeType    string
	reader      server.ResourceFunc
	tags        []string
}

// NewResource creates a new resource builder
func NewResource(uri string) *ResourceBuilder {
	return &ResourceBuilder{uri: uri}
}

// Name sets the resource name
func (rb *ResourceBuilder) Name(name string) *ResourceBuilder {
	rb.name = name
	return rb
}

// Description sets the resource description
func (rb *ResourceBuilder) Description(desc string) *ResourceBuilder {
	rb.description = desc
	return rb
}

// MimeType sets the resource MIME type
func (rb *ResourceBuilder) MimeType(mimeType string) *ResourceBuilder {
	rb.mimeType = mimeType
	return rb
}

// Reader sets the resource reader function
func (rb *ResourceBuilder) Reader(fn server.ResourceFunc) *ResourceBuilder {
	rb.reader = fn
	return rb
}

// Tags sets the resource tags
func (rb *ResourceBuilder) Tags(tags ...string) *ResourceBuilder {
	rb.tags = tags
	return rb
}

// Build creates the ResourceHandler
func (rb *ResourceBuilder) Build() *server.ResourceHandler {
	return &server.ResourceHandler{
		URI:         rb.uri,
		Name:        rb.name,
		Description: rb.description,
		MimeType:    rb.mimeType,
		Reader:      rb.reader,
		Tags:        rb.tags,
	}
}

// ResourceTemplateBuilder creates resource templates using a fluent API
type ResourceTemplateBuilder struct {
	uriTemplate string
	name        string
	description string
	mimeType    string
	reader      server.ResourceTemplateFunc
	tags        []string
}

// NewResourceTemplate creates a new resource template builder
func NewResourceTemplate(uriTemplate string) *ResourceTemplateBuilder {
	return &ResourceTemplateBuilder{uriTemplate: uriTemplate}
}

// Name sets the resource template name
func (rtb *ResourceTemplateBuilder) Name(name string) *ResourceTemplateBuilder {
	rtb.name = name
	return rtb
}

// Description sets the resource template description
func (rtb *ResourceTemplateBuilder) Description(desc string) *ResourceTemplateBuilder {
	rtb.description = desc
	return rtb
}

// MimeType sets the resource template MIME type
func (rtb *ResourceTemplateBuilder) MimeType(mimeType string) *ResourceTemplateBuilder {
	rtb.mimeType = mimeType
	return rtb
}

// Reader sets the resource template reader function
func (rtb *ResourceTemplateBuilder) Reader(fn server.ResourceTemplateFunc) *ResourceTemplateBuilder {
	rtb.reader = fn
	return rtb
}

// ReaderSimple sets a simple reader that takes a single path parameter
func (rtb *ResourceTemplateBuilder) ReaderSimple(fn func(context.Context, string) ([]byte, error)) *ResourceTemplateBuilder {
	rtb.reader = func(ctx context.Context, params map[string]string) ([]byte, error) {
		// Assume first parameter is the path
		for _, v := range params {
			return fn(ctx, v)
		}
		return nil, &server.ErrorContext{Message: "no parameters provided"}
	}
	return rtb
}

// Tags sets the resource template tags
func (rtb *ResourceTemplateBuilder) Tags(tags ...string) *ResourceTemplateBuilder {
	rtb.tags = tags
	return rtb
}

// Build creates the ResourceTemplateHandler
func (rtb *ResourceTemplateBuilder) Build() *server.ResourceTemplateHandler {
	return &server.ResourceTemplateHandler{
		URITemplate: rtb.uriTemplate,
		Name:        rtb.name,
		Description: rtb.description,
		MimeType:    rtb.mimeType,
		Reader:      rtb.reader,
		Tags:        rtb.tags,
	}
}
