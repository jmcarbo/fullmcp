package mcp

import (
	"encoding/json"
	"testing"
)

func TestRootSerialization(t *testing.T) {
	root := Root{
		URI:  "file:///home/user/projects/myproject",
		Name: "My Project",
	}

	// Serialize
	data, err := json.Marshal(root)
	if err != nil {
		t.Fatalf("failed to marshal root: %v", err)
	}

	// Deserialize
	var decoded Root
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal root: %v", err)
	}

	// Verify
	if decoded.URI != root.URI {
		t.Errorf("URI mismatch: got %s, want %s", decoded.URI, root.URI)
	}

	if decoded.Name != root.Name {
		t.Errorf("Name mismatch: got %s, want %s", decoded.Name, root.Name)
	}
}

func TestRootsListResult(t *testing.T) {
	result := RootsListResult{
		Roots: []Root{
			{
				URI:  "file:///home/user/projects/project1",
				Name: "Project 1",
			},
			{
				URI:  "file:///home/user/projects/project2",
				Name: "Project 2",
			},
		},
	}

	// Serialize
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal result: %v", err)
	}

	// Deserialize
	var decoded RootsListResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// Verify
	if len(decoded.Roots) != 2 {
		t.Errorf("roots count mismatch: got %d, want 2", len(decoded.Roots))
	}

	if decoded.Roots[0].URI != "file:///home/user/projects/project1" {
		t.Errorf("first root URI mismatch")
	}

	if decoded.Roots[1].Name != "Project 2" {
		t.Errorf("second root name mismatch")
	}
}

func TestRootsCapability(t *testing.T) {
	cap := RootsCapability{
		ListChanged: true,
	}

	// Serialize
	data, err := json.Marshal(cap)
	if err != nil {
		t.Fatalf("failed to marshal capability: %v", err)
	}

	// Deserialize
	var decoded RootsCapability
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal capability: %v", err)
	}

	// Verify
	if decoded.ListChanged != true {
		t.Errorf("listChanged mismatch: got %v, want true", decoded.ListChanged)
	}
}

func TestRootWithoutName(t *testing.T) {
	// Test that name is optional
	root := Root{
		URI: "file:///home/user/documents",
	}

	data, err := json.Marshal(root)
	if err != nil {
		t.Fatalf("failed to marshal root: %v", err)
	}

	var decoded Root
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal root: %v", err)
	}

	if decoded.URI != root.URI {
		t.Errorf("URI mismatch")
	}

	if decoded.Name != "" {
		t.Errorf("expected empty name, got %s", decoded.Name)
	}
}
