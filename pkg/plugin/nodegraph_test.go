package plugin

import (
	"context"
	"encoding/json"
	"fmt"
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

func TestBuildNodeGraphFrames_RealGraphData(t *testing.T) {
	db := uniqueDBName("nodegraph")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	// Create types and graph data.
	cmds := []string{
		"CREATE VERTEX TYPE Person",
		"CREATE EDGE TYPE Knows",
		"CREATE VERTEX Person SET name = 'Alice', age = 30",
		"CREATE VERTEX Person SET name = 'Bob', age = 25",
		"CREATE VERTEX Person SET name = 'Carol', age = 35",
		"CREATE EDGE Knows FROM (SELECT FROM Person WHERE name = 'Alice') TO (SELECT FROM Person WHERE name = 'Bob') SET since = 2019",
		"CREATE EDGE Knows FROM (SELECT FROM Person WHERE name = 'Bob') TO (SELECT FROM Person WHERE name = 'Carol') SET since = 2021",
	}
	for _, cmd := range cmds {
		if _, err := testArcadeDB.ExecuteCommand(db, "sql", cmd); err != nil {
			t.Fatalf("setup command %q failed: %v", cmd, err)
		}
	}

	// Query with graph serializer via the client.
	client := testArcadeDB.NewTestClient(db)
	respBytes, err := client.ExecuteCommand(context.Background(), &CommandRequest{
		Language:   "sql",
		Command:    "SELECT FROM Person",
		Serializer: "graph",
		Limit:      100,
	})
	if err != nil {
		t.Fatalf("graph query failed: %v", err)
	}

	var resp GraphResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("failed to unmarshal graph response: %v", err)
	}

	t.Logf("Graph response: %d vertices, %d edges", len(resp.Result.Vertices), len(resp.Result.Edges))

	if len(resp.Result.Vertices) < 3 {
		t.Fatalf("expected at least 3 vertices, got %d", len(resp.Result.Vertices))
	}

	// Pass through BuildNodeGraphFrames.
	frames, err := BuildNodeGraphFrames(resp.Result.Vertices, resp.Result.Edges)
	if err != nil {
		t.Fatalf("BuildNodeGraphFrames failed: %v", err)
	}

	if len(frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(frames))
	}

	nodesFrame := frames[0]
	edgesFrame := frames[1]

	if nodesFrame.Name != "nodes" {
		t.Errorf("expected nodes frame name 'nodes', got %q", nodesFrame.Name)
	}
	if nodesFrame.Rows() < 3 {
		t.Errorf("expected at least 3 node rows, got %d", nodesFrame.Rows())
	}

	// Verify that node titles contain our person names.
	titleField := nodesFrame.Fields[1]
	foundNames := map[string]bool{}
	for i := 0; i < nodesFrame.Rows(); i++ {
		val := titleField.At(i)
		if s, ok := val.(*string); ok && s != nil {
			foundNames[*s] = true
		}
	}
	for _, name := range []string{"Alice", "Bob", "Carol"} {
		if !foundNames[name] {
			t.Errorf("expected to find node with title %q", name)
		}
	}

	// Verify preferred visualization metadata.
	if nodesFrame.Meta == nil || nodesFrame.Meta.PreferredVisualization != "nodeGraph" {
		t.Error("nodes frame missing nodeGraph preferred visualization")
	}
	if edgesFrame.Meta == nil || edgesFrame.Meta.PreferredVisualization != "nodeGraph" {
		t.Error("edges frame missing nodeGraph preferred visualization")
	}

	t.Logf("nodes frame: %d rows, %d fields", nodesFrame.Rows(), len(nodesFrame.Fields))
	t.Logf("edges frame: %d rows, %d fields", edgesFrame.Rows(), len(edgesFrame.Fields))
}

func TestBuildNodeGraphFrames_VertexProperties(t *testing.T) {
	vertices := []GraphElement{
		{
			R: "#1:0", T: "Server",
			P: map[string]interface{}{
				"name":     "web-01",
				"cpu":      float64(85.5),
				"region":   "us-east-1",
				"active":   true,
				"priority": float64(1),
			},
			I: float64(3), O: float64(5),
		},
		{
			R: "#1:1", T: "Server",
			P: map[string]interface{}{
				"name":     "web-02",
				"cpu":      float64(42.1),
				"region":   "eu-west-1",
				"active":   false,
				"priority": float64(2),
			},
			I: float64(1), O: float64(2),
		},
	}

	frames, err := BuildNodeGraphFrames(vertices, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nodesFrame := frames[0]

	// The first 4 fields are id, title, subtitle, mainstat.
	// All remaining fields should be detail__* fields.
	expectedProps := map[string]bool{
		"name": true, "cpu": true, "region": true, "active": true, "priority": true,
	}

	foundDetailFields := map[string]bool{}
	for _, field := range nodesFrame.Fields {
		if len(field.Name) > 8 && field.Name[:8] == "detail__" {
			propName := field.Name[8:]
			foundDetailFields[propName] = true
		}
	}

	for prop := range expectedProps {
		if !foundDetailFields[prop] {
			t.Errorf("expected detail__%s field, but it was not found", prop)
		}
	}

	// Verify values for a specific detail field.
	for _, field := range nodesFrame.Fields {
		if field.Name == "detail__region" {
			val0 := field.At(0)
			val1 := field.At(1)
			if s, ok := val0.(*string); ok && s != nil {
				if *s != "us-east-1" {
					t.Errorf("expected detail__region[0] = 'us-east-1', got %q", *s)
				}
			} else {
				t.Error("expected detail__region[0] to be a non-nil string pointer")
			}
			if s, ok := val1.(*string); ok && s != nil {
				if *s != "eu-west-1" {
					t.Errorf("expected detail__region[1] = 'eu-west-1', got %q", *s)
				}
			} else {
				t.Error("expected detail__region[1] to be a non-nil string pointer")
			}
			break
		}
	}
}

func TestBuildNodeGraphFrames_EdgeProperties(t *testing.T) {
	vertices := []GraphElement{
		{R: "#1:0", T: "Node", P: map[string]interface{}{"name": "A"}, I: float64(0), O: float64(1)},
		{R: "#1:1", T: "Node", P: map[string]interface{}{"name": "B"}, I: float64(1), O: float64(0)},
	}
	edges := []GraphElement{
		{
			R: "#10:0", T: "Connection",
			P: map[string]interface{}{
				"weight":   float64(0.75),
				"protocol": "tcp",
				"latency":  float64(15),
			},
			O: "#1:0", I: "#1:1",
		},
	}

	frames, err := BuildNodeGraphFrames(vertices, edges)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	edgesFrame := frames[1]

	expectedEdgeProps := map[string]bool{
		"weight": true, "protocol": true, "latency": true,
	}

	foundDetailFields := map[string]bool{}
	for _, field := range edgesFrame.Fields {
		if len(field.Name) > 8 && field.Name[:8] == "detail__" {
			propName := field.Name[8:]
			foundDetailFields[propName] = true
		}
	}

	for prop := range expectedEdgeProps {
		if !foundDetailFields[prop] {
			t.Errorf("expected detail__%s field on edges frame, but it was not found", prop)
		}
	}

	// Verify a specific edge detail value.
	for _, field := range edgesFrame.Fields {
		if field.Name == "detail__protocol" {
			if s, ok := field.At(0).(*string); ok && s != nil {
				if *s != "tcp" {
					t.Errorf("expected detail__protocol = 'tcp', got %q", *s)
				}
			} else {
				t.Error("expected detail__protocol to be a non-nil string pointer")
			}
			break
		}
	}

	// Verify edge mainstat is the type name.
	mainstatField := edgesFrame.Fields[3]
	if s, ok := mainstatField.At(0).(*string); ok && s != nil {
		if *s != "Connection" {
			t.Errorf("expected edge mainstat 'Connection', got %q", *s)
		}
	}
}

func TestBuildNodeGraphFrames_TitleResolution(t *testing.T) {
	tests := []struct {
		desc     string
		props    map[string]interface{}
		rid      string
		expected string
	}{
		{
			desc:     "uses name when present",
			props:    map[string]interface{}{"name": "Alice", "label": "Person A", "title": "Ms. Alice"},
			rid:      "#1:0",
			expected: "Alice",
		},
		{
			desc:     "uses label when no name",
			props:    map[string]interface{}{"label": "Sensor A", "title": "Primary Sensor"},
			rid:      "#2:0",
			expected: "Sensor A",
		},
		{
			desc:     "uses title when no name or label",
			props:    map[string]interface{}{"title": "Important Node", "color": "red"},
			rid:      "#3:0",
			expected: "Important Node",
		},
		{
			desc:     "falls back to RID",
			props:    map[string]interface{}{"color": "blue", "size": float64(10)},
			rid:      "#4:0",
			expected: "#4:0",
		},
		{
			desc:     "name takes priority over label and title",
			props:    map[string]interface{}{"title": "T", "label": "L", "name": "N"},
			rid:      "#5:0",
			expected: "N",
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			v := GraphElement{R: tc.rid, T: "TestType", P: tc.props}
			title := resolveTitle(v)
			if title != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, title)
			}
		})
	}

	// Also verify title resolution works through BuildNodeGraphFrames.
	vertices := make([]GraphElement, len(tests))
	for i, tc := range tests {
		vertices[i] = GraphElement{R: tc.rid, T: "TestType", P: tc.props, I: float64(0), O: float64(0)}
	}

	frames, err := BuildNodeGraphFrames(vertices, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	titleField := frames[0].Fields[1]
	for i, tc := range tests {
		val := titleField.At(i)
		if s, ok := val.(*string); ok && s != nil {
			if *s != tc.expected {
				t.Errorf("row %d (%s): expected title %q, got %q", i, tc.desc, tc.expected, *s)
			}
		} else {
			t.Errorf("row %d (%s): expected non-nil string pointer", i, tc.desc)
		}
	}
}

func TestBuildNodeGraphFrames_NullProperties(t *testing.T) {
	vertices := []GraphElement{
		{
			R: "#1:0", T: "Item",
			P: map[string]interface{}{
				"name":        "Widget",
				"description": nil,
				"count":       nil,
				"category":    "tools",
			},
			I: float64(0), O: float64(0),
		},
		{
			R: "#1:1", T: "Item",
			P: map[string]interface{}{
				"name":        nil,
				"description": "A gadget",
				"count":       float64(5),
				"category":    nil,
			},
			I: float64(0), O: float64(0),
		},
	}

	frames, err := BuildNodeGraphFrames(vertices, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nodesFrame := frames[0]

	// Verify the frame built without panicking and has the right row count.
	if nodesFrame.Rows() != 2 {
		t.Fatalf("expected 2 rows, got %d", nodesFrame.Rows())
	}

	// Check detail fields for nil handling.
	for _, field := range nodesFrame.Fields {
		if field.Name == "detail__description" {
			// First vertex has nil description, so field value should be nil pointer.
			val0 := field.At(0)
			if val0 != (*string)(nil) {
				t.Errorf("expected detail__description[0] to be nil, got %v", val0)
			}
			// Second vertex has "A gadget".
			val1 := field.At(1)
			if s, ok := val1.(*string); ok && s != nil {
				if *s != "A gadget" {
					t.Errorf("expected detail__description[1] = 'A gadget', got %q", *s)
				}
			} else {
				t.Error("expected detail__description[1] to be non-nil")
			}
		}
		if field.Name == "detail__count" {
			// First vertex: nil count.
			val0 := field.At(0)
			if val0 != (*string)(nil) {
				t.Errorf("expected detail__count[0] to be nil, got %v", val0)
			}
			// Second vertex: count = 5.
			val1 := field.At(1)
			if s, ok := val1.(*string); ok && s != nil {
				if *s != fmt.Sprintf("%v", float64(5)) {
					t.Errorf("expected detail__count[1] = '5', got %q", *s)
				}
			} else {
				t.Error("expected detail__count[1] to be non-nil")
			}
		}
	}

	// Verify title resolution handles nil name gracefully.
	// First vertex has name="Widget", second has name=nil so falls back.
	titleField := nodesFrame.Fields[1]
	title0 := titleField.At(0).(*string)
	if *title0 != "Widget" {
		t.Errorf("expected title[0] = 'Widget', got %q", *title0)
	}
	// Second vertex: name is nil, description is present but not a title key, so falls back to RID.
	title1 := titleField.At(1).(*string)
	if *title1 != "#1:1" {
		t.Errorf("expected title[1] = '#1:1' (RID fallback), got %q", *title1)
	}
}
