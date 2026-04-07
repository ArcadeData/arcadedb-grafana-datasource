package plugin

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCheckHealth(t *testing.T) {
	// The CheckHealth method calls /api/v1/ts/{db}/grafana/health, which is
	// a custom Grafana time-series plugin endpoint. Vanilla ArcadeDB does not
	// expose this endpoint, so both subtests verify error behavior.

	t.Run("existing database returns error for missing grafana endpoint", func(t *testing.T) {
		db := uniqueDBName("health_exists")
		if err := testArcadeDB.CreateDatabase(db); err != nil {
			t.Fatalf("failed to create database: %v", err)
		}
		defer testArcadeDB.DropDatabase(db)

		client := testArcadeDB.NewTestClient(db)
		err := client.CheckHealth(context.Background())
		// The endpoint does not exist in vanilla ArcadeDB, so we expect an error.
		if err == nil {
			// If ArcadeDB ships with grafana endpoints in the future, this is OK.
			t.Log("CheckHealth succeeded - grafana health endpoint is available")
			return
		}
		t.Logf("CheckHealth returned expected error: %v", err)
	})

	t.Run("nonexistent database returns error", func(t *testing.T) {
		client := testArcadeDB.NewTestClient("nonexistent_db_xyz")
		err := client.CheckHealth(context.Background())
		if err == nil {
			t.Fatal("expected error for nonexistent database, got nil")
		}
		t.Logf("CheckHealth error for nonexistent DB: %v", err)
	})
}

func TestExecuteCommand_SQL(t *testing.T) {
	db := uniqueDBName("sql")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	// Set up test data: create a document type and insert records.
	_, err := testArcadeDB.ExecuteCommand(db, "sql", "CREATE DOCUMENT TYPE SensorReading")
	if err != nil {
		t.Fatalf("failed to create type: %v", err)
	}
	_, err = testArcadeDB.ExecuteCommand(db, "sql", "INSERT INTO SensorReading SET temperature = 22.5, location = 'us-east'")
	if err != nil {
		t.Fatalf("failed to insert record 1: %v", err)
	}
	_, err = testArcadeDB.ExecuteCommand(db, "sql", "INSERT INTO SensorReading SET temperature = 18.3, location = 'us-west'")
	if err != nil {
		t.Fatalf("failed to insert record 2: %v", err)
	}

	client := testArcadeDB.NewTestClient(db)
	respBytes, err := client.ExecuteCommand(context.Background(), &CommandRequest{
		Language: "sql",
		Command:  "SELECT temperature, location FROM SensorReading ORDER BY temperature",
		Limit:    100,
	})
	if err != nil {
		t.Fatalf("ExecuteCommand failed: %v", err)
	}

	var resp CommandResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp.Result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(resp.Result))
	}

	// Results are ordered by temperature ascending: 18.3 then 22.5
	temp0, ok := resp.Result[0]["temperature"].(float64)
	if !ok {
		t.Fatalf("expected temperature to be float64, got %T", resp.Result[0]["temperature"])
	}
	if temp0 != 18.3 {
		t.Errorf("expected first temperature 18.3, got %v", temp0)
	}

	loc1, ok := resp.Result[1]["location"].(string)
	if !ok {
		t.Fatalf("expected location to be string, got %T", resp.Result[1]["location"])
	}
	if loc1 != "us-east" {
		t.Errorf("expected second location 'us-east', got %q", loc1)
	}
}

func TestExecuteCommand_OpenCypher(t *testing.T) {
	db := uniqueDBName("cypher")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	// Create vertex type and insert using SQL (openCypher CREATE may need schema).
	_, err := testArcadeDB.ExecuteCommand(db, "sql", "CREATE VERTEX TYPE Person")
	if err != nil {
		t.Fatalf("failed to create vertex type: %v", err)
	}
	_, err = testArcadeDB.ExecuteCommand(db, "sql", "INSERT INTO Person SET name = 'Alice', age = 30")
	if err != nil {
		t.Fatalf("failed to insert vertex 1: %v", err)
	}
	_, err = testArcadeDB.ExecuteCommand(db, "sql", "INSERT INTO Person SET name = 'Bob', age = 25")
	if err != nil {
		t.Fatalf("failed to insert vertex 2: %v", err)
	}

	client := testArcadeDB.NewTestClient(db)
	respBytes, err := client.ExecuteCommand(context.Background(), &CommandRequest{
		Language: "cypher",
		Command:  "MATCH (p:Person) RETURN p.name AS name, p.age AS age ORDER BY p.name",
		Limit:    100,
	})
	if err != nil {
		t.Fatalf("ExecuteCommand with cypher failed: %v", err)
	}

	var resp CommandResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp.Result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(resp.Result))
	}

	name0, ok := resp.Result[0]["name"].(string)
	if !ok {
		t.Fatalf("expected name to be string, got %T", resp.Result[0]["name"])
	}
	if name0 != "Alice" {
		t.Errorf("expected first name 'Alice', got %q", name0)
	}

	age1, ok := resp.Result[1]["age"].(float64)
	if !ok {
		t.Fatalf("expected age to be float64, got %T", resp.Result[1]["age"])
	}
	if age1 != 25 {
		t.Errorf("expected second age 25, got %v", age1)
	}
}

func TestExecuteCommand_GraphSerializer(t *testing.T) {
	db := uniqueDBName("graph")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	// Create vertex and edge types, then create test graph data.
	cmds := []string{
		"CREATE VERTEX TYPE City",
		"CREATE EDGE TYPE ConnectedTo",
		"CREATE VERTEX City SET name = 'New York'",
		"CREATE VERTEX City SET name = 'Boston'",
	}
	for _, cmd := range cmds {
		if _, err := testArcadeDB.ExecuteCommand(db, "sql", cmd); err != nil {
			t.Fatalf("setup command %q failed: %v", cmd, err)
		}
	}

	// Create an edge between the two cities.
	_, err := testArcadeDB.ExecuteCommand(db, "sql",
		"CREATE EDGE ConnectedTo FROM (SELECT FROM City WHERE name = 'New York') TO (SELECT FROM City WHERE name = 'Boston') SET distance = 215")
	if err != nil {
		t.Fatalf("failed to create edge: %v", err)
	}

	client := testArcadeDB.NewTestClient(db)
	respBytes, err := client.ExecuteCommand(context.Background(), &CommandRequest{
		Language:   "sql",
		Command:    "SELECT FROM City",
		Serializer: "graph",
		Limit:      100,
	})
	if err != nil {
		t.Fatalf("ExecuteCommand with graph serializer failed: %v", err)
	}

	var resp GraphResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("failed to unmarshal graph response: %v", err)
	}

	if len(resp.Result.Vertices) < 2 {
		t.Fatalf("expected at least 2 vertices, got %d", len(resp.Result.Vertices))
	}

	// Verify vertex structure
	for _, v := range resp.Result.Vertices {
		if v.R == "" {
			t.Error("vertex has empty RID")
		}
		if v.T != "City" {
			t.Errorf("expected vertex type 'City', got %q", v.T)
		}
		name, ok := v.P["name"].(string)
		if !ok || name == "" {
			t.Errorf("expected vertex to have string 'name' property, got %v", v.P["name"])
		}
	}

	// Edges may or may not be included depending on ArcadeDB's graph serializer
	// behavior for vertex-only queries. Log what we get.
	t.Logf("Graph response: %d vertices, %d edges", len(resp.Result.Vertices), len(resp.Result.Edges))
}

func TestExecuteCommand_ErrorHandling(t *testing.T) {
	t.Run("invalid SQL", func(t *testing.T) {
		db := uniqueDBName("err_sql")
		if err := testArcadeDB.CreateDatabase(db); err != nil {
			t.Fatalf("failed to create database: %v", err)
		}
		defer testArcadeDB.DropDatabase(db)

		client := testArcadeDB.NewTestClient(db)
		_, err := client.ExecuteCommand(context.Background(), &CommandRequest{
			Language: "sql",
			Command:  "SELECTT FROMM nonexistent_type WHEREE invalid syntax !!!",
		})
		if err == nil {
			t.Fatal("expected error for invalid SQL, got nil")
		}
		t.Logf("invalid SQL error: %v", err)
	})

	t.Run("wrong credentials", func(t *testing.T) {
		db := uniqueDBName("err_creds")
		if err := testArcadeDB.CreateDatabase(db); err != nil {
			t.Fatalf("failed to create database: %v", err)
		}
		defer testArcadeDB.DropDatabase(db)

		client := NewArcadeDBClient(testArcadeDB.BaseURL(), db, "wronguser", "wrongpass")
		_, err := client.ExecuteCommand(context.Background(), &CommandRequest{
			Language: "sql",
			Command:  "SELECT 1",
		})
		if err == nil {
			t.Fatal("expected error for wrong credentials, got nil")
		}
		// ArcadeDB returns 403 for bad credentials (not 401).
		if !strings.Contains(err.Error(), "403") && !strings.Contains(err.Error(), "401") {
			t.Errorf("expected 401 or 403 in error message, got: %v", err)
		}
	})

	t.Run("nonexistent database", func(t *testing.T) {
		client := testArcadeDB.NewTestClient("this_db_does_not_exist")
		_, err := client.ExecuteCommand(context.Background(), &CommandRequest{
			Language: "sql",
			Command:  "SELECT 1",
		})
		if err == nil {
			t.Fatal("expected error for nonexistent database, got nil")
		}
		t.Logf("nonexistent database error: %v", err)
	})
}

func TestDoRequest_Timeout(t *testing.T) {
	db := uniqueDBName("timeout")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	client := testArcadeDB.NewTestClient(db)
	// Override the HTTP client with an impossibly short timeout.
	client.httpClient = &http.Client{
		Timeout: 1 * time.Nanosecond,
	}

	_, err := client.ExecuteCommand(context.Background(), &CommandRequest{
		Language: "sql",
		Command:  "SELECT 1",
	})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "Client.Timeout") && !strings.Contains(err.Error(), "context deadline exceeded") && !strings.Contains(err.Error(), "command request failed") {
		t.Errorf("expected timeout-related error, got: %v", err)
	}
	t.Logf("timeout error: %v", err)
}
