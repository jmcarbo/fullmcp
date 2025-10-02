package client

import (
	"context"

	"github.com/jmcarbo/fullmcp/mcp"
)

// GetCompletion requests completion suggestions from the server
func (c *Client) GetCompletion(ctx context.Context, ref mcp.CompletionRef, arg mcp.CompletionArgument) ([]string, error) {
	params := mcp.CompleteRequest{
		Ref:      ref,
		Argument: arg,
	}

	var result mcp.CompleteResult
	if err := c.call(ctx, "completion/complete", params, &result); err != nil {
		return nil, err
	}

	return result.Completion.Values, nil
}
