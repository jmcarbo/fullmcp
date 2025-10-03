package server

import (
	"encoding/json"
	"testing"

	"github.com/jmcarbo/fullmcp/mcp"
)

func TestConvertToContent(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		wantLen     int
		wantType    string
		wantText    string
		wantError   bool
		checkMime   bool
		wantMime    string
	}{
		{
			name:     "nil input",
			input:    nil,
			wantLen:  1,
			wantType: "text",
			wantText: "",
		},
		{
			name: "TextContent",
			input: mcp.TextContent{
				Type: "text",
				Text: "hello",
			},
			wantLen:  1,
			wantType: "text",
			wantText: "hello",
		},
		{
			name: "ImageContent",
			input: mcp.ImageContent{
				Type:     "image",
				Data:     "base64data",
				MimeType: "image/png",
			},
			wantLen:   1,
			wantType:  "image",
			checkMime: true,
			wantMime:  "image/png",
		},
		{
			name: "AudioContent",
			input: mcp.AudioContent{
				Type:     "audio",
				Data:     "base64data",
				MimeType: "audio/mp3",
			},
			wantLen:   1,
			wantType:  "audio",
			checkMime: true,
			wantMime:  "audio/mp3",
		},
		{
			name: "ResourceContent",
			input: mcp.ResourceContent{
				Type:     "resource",
				URI:      "file:///test.txt",
				MimeType: "text/plain",
				Text:     "content",
			},
			wantLen:  1,
			wantType: "resource",
		},
		{
			name: "slice of Content",
			input: []mcp.Content{
				mcp.TextContent{Type: "text", Text: "first"},
				mcp.TextContent{Type: "text", Text: "second"},
			},
			wantLen: 2,
		},
		{
			name:     "string input",
			input:    "simple string",
			wantLen:  1,
			wantType: "text",
			wantText: "simple string",
		},
		{
			name:     "byte slice input",
			input:    []byte("byte data"),
			wantLen:  1,
			wantType: "text",
			wantText: "byte data",
		},
		{
			name:     "integer input",
			input:    42,
			wantLen:  1,
			wantType: "text",
			wantText: "42",
		},
		{
			name:     "float input",
			input:    3.14,
			wantLen:  1,
			wantType: "text",
			wantText: "3.14",
		},
		{
			name: "struct input",
			input: struct {
				Name  string `json:"name"`
				Value int    `json:"value"`
			}{
				Name:  "test",
				Value: 123,
			},
			wantLen:  1,
			wantType: "text",
			wantText: `{"name":"test","value":123}`,
		},
		{
			name: "map input",
			input: map[string]interface{}{
				"key": "value",
				"num": 42,
			},
			wantLen:  1,
			wantType: "text",
			// JSON order is not guaranteed, so we'll check differently
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToContent(tt.input)

			if (err != nil) != tt.wantError {
				t.Errorf("convertToContent() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if len(got) != tt.wantLen {
				t.Errorf("convertToContent() got %d items, want %d", len(got), tt.wantLen)
				return
			}

			if tt.wantLen > 0 && tt.wantType != "" {
				if got[0].ContentType() != tt.wantType {
					t.Errorf("convertToContent() type = %v, want %v", got[0].ContentType(), tt.wantType)
				}
			}

			if tt.wantText != "" {
				if textContent, ok := got[0].(mcp.TextContent); ok {
					if textContent.Text != tt.wantText {
						t.Errorf("convertToContent() text = %v, want %v", textContent.Text, tt.wantText)
					}
				}
			}

			if tt.checkMime {
				switch v := got[0].(type) {
				case mcp.ImageContent:
					if v.MimeType != tt.wantMime {
						t.Errorf("convertToContent() mimeType = %v, want %v", v.MimeType, tt.wantMime)
					}
				case mcp.AudioContent:
					if v.MimeType != tt.wantMime {
						t.Errorf("convertToContent() mimeType = %v, want %v", v.MimeType, tt.wantMime)
					}
				}
			}

			// Special handling for map test - check if it's valid JSON
			if tt.name == "map input" {
				if textContent, ok := got[0].(mcp.TextContent); ok {
					var result map[string]interface{}
					if err := json.Unmarshal([]byte(textContent.Text), &result); err != nil {
						t.Errorf("convertToContent() returned invalid JSON: %v", err)
					}
				}
			}
		})
	}
}

func TestConvertToContent_PreservesContentTypes(t *testing.T) {
	// Test that different content types are preserved correctly
	contents := []mcp.Content{
		mcp.TextContent{Type: "text", Text: "text"},
		mcp.ImageContent{Type: "image", Data: "img", MimeType: "image/png"},
		mcp.AudioContent{Type: "audio", Data: "audio", MimeType: "audio/mp3"},
	}

	result, err := convertToContent(contents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != len(contents) {
		t.Fatalf("expected %d items, got %d", len(contents), len(result))
	}

	// Verify each type is preserved
	if _, ok := result[0].(mcp.TextContent); !ok {
		t.Error("first item should be TextContent")
	}
	if _, ok := result[1].(mcp.ImageContent); !ok {
		t.Error("second item should be ImageContent")
	}
	if _, ok := result[2].(mcp.AudioContent); !ok {
		t.Error("third item should be AudioContent")
	}
}
