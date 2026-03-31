package plugin

import (
	"testing"
)

func TestBuildNodeGraphFrames_Basic(t *testing.T) {
	vertices := []GraphElement{
		{R: "#12:0", T: "Person", P: map[string]interface{}{"name": "Alice", "age": float64(30)}, I: float64(2), O: float64(3)},
		{R: "#12:1", T: "Person", P: map[string]interface{}{"name": "Bob", "age": float64(25)}, I: float64(1), O: float64(2)},
	}
	edges := []GraphElement{
		{R: "#20:0", T: "FriendOf", P: map[string]interface{}{"since": float64(2020)}, I: "#12:1", O: "#12:0"},
	}

	frames, err := BuildNodeGraphFrames(vertices, edges)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(frames))
	}

	nodesFrame := frames[0]
	edgesFrame := frames[1]

	// Verify nodes frame
	if nodesFrame.Name != "nodes" {
		t.Errorf("expected nodes frame name 'nodes', got %q", nodesFrame.Name)
	}
	if nodesFrame.Rows() != 2 {
		t.Errorf("expected 2 node rows, got %d", nodesFrame.Rows())
	}

	// Check node IDs
	idField := nodesFrame.Fields[0]
	if *idField.At(0).(*string) != "#12:0" {
		t.Errorf("expected node 0 id '#12:0', got %q", *idField.At(0).(*string))
	}

	// Check titles (should use 'name' property)
	titleField := nodesFrame.Fields[1]
	if *titleField.At(0).(*string) != "Alice" {
		t.Errorf("expected node 0 title 'Alice', got %q", *titleField.At(0).(*string))
	}
	if *titleField.At(1).(*string) != "Bob" {
		t.Errorf("expected node 1 title 'Bob', got %q", *titleField.At(1).(*string))
	}

	// Verify edges frame
	if edgesFrame.Name != "edges" {
		t.Errorf("expected edges frame name 'edges', got %q", edgesFrame.Name)
	}
	if edgesFrame.Rows() != 1 {
		t.Errorf("expected 1 edge row, got %d", edgesFrame.Rows())
	}

	// Check source (o = OUT vertex) and target (i = IN vertex)
	sourceField := edgesFrame.Fields[1]
	targetField := edgesFrame.Fields[2]
	if *sourceField.At(0).(*string) != "#12:0" {
		t.Errorf("expected edge source '#12:0', got %q", *sourceField.At(0).(*string))
	}
	if *targetField.At(0).(*string) != "#12:1" {
		t.Errorf("expected edge target '#12:1', got %q", *targetField.At(0).(*string))
	}

	// Check edge mainstat (type name)
	mainstatField := edgesFrame.Fields[3]
	if *mainstatField.At(0).(*string) != "FriendOf" {
		t.Errorf("expected edge mainstat 'FriendOf', got %q", *mainstatField.At(0).(*string))
	}
}

func TestBuildNodeGraphFrames_EmptyInput(t *testing.T) {
	frames, err := BuildNodeGraphFrames(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(frames))
	}

	if frames[0].Rows() != 0 {
		t.Errorf("expected 0 nodes, got %d", frames[0].Rows())
	}
	if frames[1].Rows() != 0 {
		t.Errorf("expected 0 edges, got %d", frames[1].Rows())
	}
}

func TestResolveTitle_FallbackToRID(t *testing.T) {
	v := GraphElement{R: "#5:0", T: "Device", P: map[string]interface{}{"serial": "ABC123"}}
	title := resolveTitle(v)
	if title != "#5:0" {
		t.Errorf("expected '#5:0', got %q", title)
	}
}

func TestResolveTitle_UsesLabel(t *testing.T) {
	v := GraphElement{R: "#5:0", T: "Device", P: map[string]interface{}{"label": "Sensor A"}}
	title := resolveTitle(v)
	if title != "Sensor A" {
		t.Errorf("expected 'Sensor A', got %q", title)
	}
}
