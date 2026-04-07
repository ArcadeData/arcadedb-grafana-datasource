package plugin

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// newTestQuery builds a backend.DataQuery and QueryModel for use in command handler tests.
func newTestQuery(rawQuery string, mode string, nodeGraph bool) (backend.DataQuery, QueryModel) {
	qm := QueryModel{
		QueryMode:        mode,
		RawQuery:         rawQuery,
		NodeGraphEnabled: nodeGraph,
	}
	qmJSON, _ := json.Marshal(qm)
	dq := backend.DataQuery{
		RefID:         "A",
		JSON:          qmJSON,
		TimeRange:     backend.TimeRange{From: time.UnixMilli(0), To: time.UnixMilli(999999999999)},
		Interval:      60 * time.Second,
		MaxDataPoints: 1000,
	}
	return dq, qm
}

func TestHandleCommandQuery_SQL(t *testing.T) {
	db := uniqueDBName("cmd_sql")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	// Create a document type with mixed field types.
	cmds := []string{
		"CREATE DOCUMENT TYPE MixedRecord",
		"INSERT INTO MixedRecord SET name = 'Alice', score = 95.5, active = true",
		"INSERT INTO MixedRecord SET name = 'Bob', score = 82.0, active = false",
	}
	for _, cmd := range cmds {
		if _, err := testArcadeDB.ExecuteCommand(db, "sql", cmd); err != nil {
			t.Fatalf("setup command %q failed: %v", cmd, err)
		}
	}

	client := testArcadeDB.NewTestClient(db)
	dq, qm := newTestQuery("SELECT name, score, active FROM MixedRecord ORDER BY name", "sql", false)

	frames, err := HandleCommandQuery(context.Background(), client, dq, qm, "sql")
	if err != nil {
		t.Fatalf("HandleCommandQuery failed: %v", err)
	}

	if len(frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(frames))
	}

	frame := frames[0]
	if frame.Name != "result" {
		t.Errorf("expected frame name 'result', got %q", frame.Name)
	}

	if len(frame.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(frame.Fields))
	}

	// Verify field types: name=string, score=number, active=boolean
	fieldTypes := map[string]string{}
	for _, f := range frame.Fields {
		fieldTypes[f.Name] = f.Type().String()
	}

	if ft, ok := fieldTypes["name"]; !ok || ft != "[]*string" {
		t.Errorf("expected field 'name' to be 'nullable string', got %q", ft)
	}
	if ft, ok := fieldTypes["score"]; !ok || ft != "[]*float64" {
		t.Errorf("expected field 'score' to be 'nullable float64', got %q", ft)
	}
	if ft, ok := fieldTypes["active"]; !ok || ft != "[]*bool" {
		t.Errorf("expected field 'active' to be 'nullable bool', got %q", ft)
	}

	// Verify row count
	if frame.Rows() != 2 {
		t.Fatalf("expected 2 rows, got %d", frame.Rows())
	}

	// Verify values (ordered by name: Alice, Bob)
	nameField := findField(frame, "name")
	scoreField := findField(frame, "score")
	activeField := findField(frame, "active")

	assertStringFieldValue(t, nameField, 0, "Alice")
	assertStringFieldValue(t, nameField, 1, "Bob")
	assertFloat64FieldValue(t, scoreField, 0, 95.5)
	assertFloat64FieldValue(t, scoreField, 1, 82.0)
	assertBoolFieldValue(t, activeField, 0, true)
	assertBoolFieldValue(t, activeField, 1, false)
}

func TestHandleCommandQuery_OpenCypher(t *testing.T) {
	db := uniqueDBName("cmd_cypher")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	cmds := []string{
		"CREATE VERTEX TYPE Item",
		"INSERT INTO Item SET name = 'Widget', price = 12.99, inStock = true",
		"INSERT INTO Item SET name = 'Gadget', price = 29.99, inStock = false",
	}
	for _, cmd := range cmds {
		if _, err := testArcadeDB.ExecuteCommand(db, "sql", cmd); err != nil {
			t.Fatalf("setup command %q failed: %v", cmd, err)
		}
	}

	client := testArcadeDB.NewTestClient(db)
	dq, qm := newTestQuery("MATCH (i:Item) RETURN i.name AS name, i.price AS price, i.inStock AS inStock ORDER BY i.name", "cypher", false)

	frames, err := HandleCommandQuery(context.Background(), client, dq, qm, "cypher")
	if err != nil {
		t.Fatalf("HandleCommandQuery failed: %v", err)
	}

	if len(frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(frames))
	}

	frame := frames[0]
	if frame.Rows() != 2 {
		t.Fatalf("expected 2 rows, got %d", frame.Rows())
	}

	// Verify field types
	fieldTypes := map[string]string{}
	for _, f := range frame.Fields {
		fieldTypes[f.Name] = f.Type().String()
	}

	if ft := fieldTypes["name"]; ft != "[]*string" {
		t.Errorf("expected 'name' to be 'nullable string', got %q", ft)
	}
	if ft := fieldTypes["price"]; ft != "[]*float64" {
		t.Errorf("expected 'price' to be 'nullable float64', got %q", ft)
	}
	if ft := fieldTypes["inStock"]; ft != "[]*bool" {
		t.Errorf("expected 'inStock' to be 'nullable bool', got %q", ft)
	}

	// Verify values
	nameField := findField(frame, "name")
	assertStringFieldValue(t, nameField, 0, "Gadget")
	assertStringFieldValue(t, nameField, 1, "Widget")
}

func TestHandleCommandQuery_Gremlin(t *testing.T) {
	db := uniqueDBName("cmd_gremlin")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	cmds := []string{
		"CREATE VERTEX TYPE Product",
		"INSERT INTO Product SET name = 'Phone', weight = 0.5, available = true",
		"INSERT INTO Product SET name = 'Laptop', weight = 2.5, available = false",
	}
	for _, cmd := range cmds {
		if _, err := testArcadeDB.ExecuteCommand(db, "sql", cmd); err != nil {
			t.Fatalf("setup command %q failed: %v", cmd, err)
		}
	}

	client := testArcadeDB.NewTestClient(db)
	dq, qm := newTestQuery("g.V().hasLabel('Product').order().by('name', asc)", "gremlin", false)

	frames, err := HandleCommandQuery(context.Background(), client, dq, qm, "gremlin")
	if err != nil {
		t.Fatalf("HandleCommandQuery failed: %v", err)
	}

	if len(frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(frames))
	}

	frame := frames[0]
	if frame.Rows() != 2 {
		t.Fatalf("expected 2 rows, got %d", frame.Rows())
	}

	// Gremlin returns full vertex records with properties - verify we got data
	t.Logf("Gremlin result: %d fields, %d rows", len(frame.Fields), frame.Rows())
}

func TestHandleCommandQuery_GraphMode(t *testing.T) {
	db := uniqueDBName("cmd_graph")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	cmds := []string{
		"CREATE VERTEX TYPE Person",
		"CREATE EDGE TYPE Knows",
		"CREATE VERTEX Person SET name = 'Alice', age = 30",
		"CREATE VERTEX Person SET name = 'Bob', age = 25",
		"CREATE EDGE Knows FROM (SELECT FROM Person WHERE name = 'Alice') TO (SELECT FROM Person WHERE name = 'Bob') SET since = 2020",
	}
	for _, cmd := range cmds {
		if _, err := testArcadeDB.ExecuteCommand(db, "sql", cmd); err != nil {
			t.Fatalf("setup command %q failed: %v", cmd, err)
		}
	}

	client := testArcadeDB.NewTestClient(db)
	// nodeGraph: true with cypher language triggers graph serializer
	dq, qm := newTestQuery("MATCH (p:Person)-[k:Knows]->(p2:Person) RETURN p, k, p2", "cypher", true)

	frames, err := HandleCommandQuery(context.Background(), client, dq, qm, "cypher")
	if err != nil {
		t.Fatalf("HandleCommandQuery failed: %v", err)
	}

	if len(frames) != 2 {
		t.Fatalf("expected 2 frames (nodes + edges), got %d", len(frames))
	}

	nodesFrame := frames[0]
	edgesFrame := frames[1]

	if nodesFrame.Name != "nodes" {
		t.Errorf("expected first frame name 'nodes', got %q", nodesFrame.Name)
	}
	if edgesFrame.Name != "edges" {
		t.Errorf("expected second frame name 'edges', got %q", edgesFrame.Name)
	}

	// Verify PreferredVisualization is set
	if nodesFrame.Meta == nil || nodesFrame.Meta.PreferredVisualization != "nodeGraph" {
		t.Error("expected nodes frame PreferredVisualization to be 'nodeGraph'")
	}
	if edgesFrame.Meta == nil || edgesFrame.Meta.PreferredVisualization != "nodeGraph" {
		t.Error("expected edges frame PreferredVisualization to be 'nodeGraph'")
	}

	// Verify nodes frame has required fields
	nodeFieldNames := map[string]bool{}
	for _, f := range nodesFrame.Fields {
		nodeFieldNames[f.Name] = true
	}
	for _, required := range []string{"id", "title", "subtitle", "mainstat"} {
		if !nodeFieldNames[required] {
			t.Errorf("nodes frame missing required field %q", required)
		}
	}

	// Verify edges frame has required fields
	edgeFieldNames := map[string]bool{}
	for _, f := range edgesFrame.Fields {
		edgeFieldNames[f.Name] = true
	}
	for _, required := range []string{"id", "source", "target", "mainstat"} {
		if !edgeFieldNames[required] {
			t.Errorf("edges frame missing required field %q", required)
		}
	}

	// Verify we got vertices and edges
	if nodesFrame.Rows() < 2 {
		t.Errorf("expected at least 2 nodes, got %d", nodesFrame.Rows())
	}
	if edgesFrame.Rows() < 1 {
		t.Errorf("expected at least 1 edge, got %d", edgesFrame.Rows())
	}

	t.Logf("Graph mode: %d nodes, %d edges", nodesFrame.Rows(), edgesFrame.Rows())
}

func TestHandleCommandQuery_EmptyResult(t *testing.T) {
	db := uniqueDBName("cmd_empty")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	_, err := testArcadeDB.ExecuteCommand(db, "sql", "CREATE DOCUMENT TYPE EmptyTable")
	if err != nil {
		t.Fatalf("failed to create type: %v", err)
	}

	client := testArcadeDB.NewTestClient(db)
	dq, qm := newTestQuery("SELECT FROM EmptyTable", "sql", false)

	frames, err := HandleCommandQuery(context.Background(), client, dq, qm, "sql")
	if err != nil {
		t.Fatalf("HandleCommandQuery failed: %v", err)
	}

	if len(frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(frames))
	}

	frame := frames[0]
	if frame.Rows() != 0 {
		t.Errorf("expected 0 rows, got %d", frame.Rows())
	}
	if len(frame.Fields) != 0 {
		t.Errorf("expected 0 fields for empty result, got %d", len(frame.Fields))
	}
}

func TestHandleCommandQuery_NullValues(t *testing.T) {
	db := uniqueDBName("cmd_null")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	// Insert records with some fields missing to produce null values.
	cmds := []string{
		"CREATE DOCUMENT TYPE NullTest",
		"INSERT INTO NullTest SET name = 'Full', score = 100.0, active = true",
		"INSERT INTO NullTest SET name = 'Partial'",
		"INSERT INTO NullTest SET score = 50.0",
	}
	for _, cmd := range cmds {
		if _, err := testArcadeDB.ExecuteCommand(db, "sql", cmd); err != nil {
			t.Fatalf("setup command %q failed: %v", cmd, err)
		}
	}

	client := testArcadeDB.NewTestClient(db)
	dq, qm := newTestQuery("SELECT name, score, active FROM NullTest ORDER BY name", "sql", false)

	frames, err := HandleCommandQuery(context.Background(), client, dq, qm, "sql")
	if err != nil {
		t.Fatalf("HandleCommandQuery failed: %v", err)
	}

	if len(frames) != 1 {
		t.Fatalf("expected 1 frame, got %d", len(frames))
	}

	frame := frames[0]
	if frame.Rows() != 3 {
		t.Fatalf("expected 3 rows, got %d", frame.Rows())
	}

	// All fields should be nullable types
	for _, f := range frame.Fields {
		typeStr := f.Type().String()
		if typeStr != "[]*string" && typeStr != "[]*float64" && typeStr != "[]*bool" {
			t.Errorf("field %q has unexpected type %q, expected a pointer-slice type", f.Name, typeStr)
		}
	}

	// Verify null handling: the record with only score=50.0 should have nil name
	nameField := findField(frame, "name")
	scoreField := findField(frame, "score")

	if nameField == nil {
		t.Fatal("missing 'name' field")
	}
	if scoreField == nil {
		t.Fatal("missing 'score' field")
	}

	// Check that at least one row has a nil value for name and at least one for score
	hasNilName := false
	hasNilScore := false
	for i := 0; i < frame.Rows(); i++ {
		v := nameField.At(i)
		if v == (*string)(nil) {
			hasNilName = true
		}
		sv := scoreField.At(i)
		if sv == (*float64)(nil) {
			hasNilScore = true
		}
	}

	if !hasNilName {
		t.Error("expected at least one nil value in 'name' field")
	}
	if !hasNilScore {
		t.Error("expected at least one nil value in 'score' field")
	}
}

func TestParseRecordResponse_FieldTypeDetection(t *testing.T) {
	db := uniqueDBName("cmd_types")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	// Insert records with various types to test buildField type inference.
	cmds := []string{
		"CREATE DOCUMENT TYPE TypeTest",
		"INSERT INTO TypeTest SET strVal = 'hello', numVal = 42.0, boolVal = true, intLike = 7",
		"INSERT INTO TypeTest SET strVal = 'world', numVal = 3.14, boolVal = false, intLike = 99",
	}
	for _, cmd := range cmds {
		if _, err := testArcadeDB.ExecuteCommand(db, "sql", cmd); err != nil {
			t.Fatalf("setup command %q failed: %v", cmd, err)
		}
	}

	client := testArcadeDB.NewTestClient(db)
	dq, qm := newTestQuery("SELECT strVal, numVal, boolVal, intLike FROM TypeTest", "sql", false)

	frames, err := HandleCommandQuery(context.Background(), client, dq, qm, "sql")
	if err != nil {
		t.Fatalf("HandleCommandQuery failed: %v", err)
	}

	frame := frames[0]

	// Build a map of field name to type string
	fieldTypes := map[string]string{}
	for _, f := range frame.Fields {
		fieldTypes[f.Name] = f.Type().String()
	}

	// strVal should be detected as string
	if ft := fieldTypes["strVal"]; ft != "[]*string" {
		t.Errorf("expected strVal to be 'nullable string', got %q", ft)
	}

	// numVal should be detected as float64 (JSON numbers deserialize as float64)
	if ft := fieldTypes["numVal"]; ft != "[]*float64" {
		t.Errorf("expected numVal to be 'nullable float64', got %q", ft)
	}

	// boolVal should be detected as bool
	if ft := fieldTypes["boolVal"]; ft != "[]*bool" {
		t.Errorf("expected boolVal to be 'nullable bool', got %q", ft)
	}

	// intLike should also be float64 (JSON numbers are always float64 in Go)
	if ft := fieldTypes["intLike"]; ft != "[]*float64" {
		t.Errorf("expected intLike to be 'nullable float64', got %q", ft)
	}
}

// --- Test helper functions ---

func findField(frame *data.Frame, name string) *data.Field {
	for _, f := range frame.Fields {
		if f.Name == name {
			return f
		}
	}
	return nil
}

func assertStringFieldValue(t *testing.T, field *data.Field, idx int, expected string) {
	t.Helper()
	if field == nil {
		t.Fatalf("field is nil")
	}
	v := field.At(idx)
	ptr, ok := v.(*string)
	if !ok {
		t.Fatalf("expected *string at index %d, got %T", idx, v)
	}
	if ptr == nil {
		t.Fatalf("expected non-nil string at index %d", idx)
	}
	if *ptr != expected {
		t.Errorf("expected %q at index %d, got %q", expected, idx, *ptr)
	}
}

func assertFloat64FieldValue(t *testing.T, field *data.Field, idx int, expected float64) {
	t.Helper()
	if field == nil {
		t.Fatalf("field is nil")
	}
	v := field.At(idx)
	ptr, ok := v.(*float64)
	if !ok {
		t.Fatalf("expected *float64 at index %d, got %T", idx, v)
	}
	if ptr == nil {
		t.Fatalf("expected non-nil float64 at index %d", idx)
	}
	if *ptr != expected {
		t.Errorf("expected %v at index %d, got %v", expected, idx, *ptr)
	}
}

func assertBoolFieldValue(t *testing.T, field *data.Field, idx int, expected bool) {
	t.Helper()
	if field == nil {
		t.Fatalf("field is nil")
	}
	v := field.At(idx)
	ptr, ok := v.(*bool)
	if !ok {
		t.Fatalf("expected *bool at index %d, got %T", idx, v)
	}
	if ptr == nil {
		t.Fatalf("expected non-nil bool at index %d", idx)
	}
	if *ptr != expected {
		t.Errorf("expected %v at index %d, got %v", expected, idx, *ptr)
	}
}
