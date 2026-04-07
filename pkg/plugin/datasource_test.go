package plugin

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// newTestDatasource creates a Datasource backed by the shared test ArcadeDB container.
func newTestDatasource(t *testing.T, db string) *Datasource {
	t.Helper()
	return &Datasource{client: testArcadeDB.NewTestClient(db)}
}

// makeQuery builds a backend.DataQuery from a refID and QueryModel.
func makeQuery(refID string, qm QueryModel) backend.DataQuery {
	qmJSON, _ := json.Marshal(qm)
	return backend.DataQuery{
		RefID:         refID,
		JSON:          qmJSON,
		TimeRange:     backend.TimeRange{From: time.UnixMilli(0), To: time.UnixMilli(999999999999)},
		Interval:      60 * time.Second,
		MaxDataPoints: 1000,
	}
}

// callResourceResponseSenderFunc adapts a function into a CallResourceResponseSender.
type callResourceResponseSenderFunc func(*backend.CallResourceResponse) error

func (f callResourceResponseSenderFunc) Send(resp *backend.CallResourceResponse) error {
	return f(resp)
}

func TestQueryData_RoutesToSQL(t *testing.T) {
	db := uniqueDBName("ds_sql")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	cmds := []string{
		"CREATE DOCUMENT TYPE Sensor",
		"INSERT INTO Sensor SET name = 'temp', value = 22.5",
	}
	for _, cmd := range cmds {
		if _, err := testArcadeDB.ExecuteCommand(db, "sql", cmd); err != nil {
			t.Fatalf("setup command %q failed: %v", cmd, err)
		}
	}

	ds := newTestDatasource(t, db)
	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{
			makeQuery("A", QueryModel{
				QueryMode: "sql",
				RawQuery:  "SELECT name, value FROM Sensor",
			}),
		},
	})
	if err != nil {
		t.Fatalf("QueryData returned top-level error: %v", err)
	}

	r, ok := resp.Responses["A"]
	if !ok {
		t.Fatal("expected response for refID 'A'")
	}
	if r.Error != nil {
		t.Fatalf("expected no error for refID 'A', got: %v", r.Error)
	}
	if len(r.Frames) == 0 {
		t.Fatal("expected at least one frame in SQL response")
	}
	if r.Frames[0].Rows() != 1 {
		t.Errorf("expected 1 row, got %d", r.Frames[0].Rows())
	}
}

func TestQueryData_RoutesToOpenCypher(t *testing.T) {
	db := uniqueDBName("ds_cypher")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	cmds := []string{
		"CREATE VERTEX TYPE Animal",
		"INSERT INTO Animal SET name = 'Cat', legs = 4",
	}
	for _, cmd := range cmds {
		if _, err := testArcadeDB.ExecuteCommand(db, "sql", cmd); err != nil {
			t.Fatalf("setup command %q failed: %v", cmd, err)
		}
	}

	ds := newTestDatasource(t, db)
	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{
			makeQuery("B", QueryModel{
				QueryMode: "cypher",
				RawQuery:  "MATCH (a:Animal) RETURN a.name AS name, a.legs AS legs",
			}),
		},
	})
	if err != nil {
		t.Fatalf("QueryData returned top-level error: %v", err)
	}

	r, ok := resp.Responses["B"]
	if !ok {
		t.Fatal("expected response for refID 'B'")
	}
	if r.Error != nil {
		t.Fatalf("expected no error for refID 'B', got: %v", r.Error)
	}
	if len(r.Frames) == 0 {
		t.Fatal("expected at least one frame in Cypher response")
	}
	if r.Frames[0].Rows() != 1 {
		t.Errorf("expected 1 row, got %d", r.Frames[0].Rows())
	}
}

func TestQueryData_RoutesToGremlin(t *testing.T) {
	db := uniqueDBName("ds_gremlin")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	cmds := []string{
		"CREATE VERTEX TYPE Color",
		"INSERT INTO Color SET name = 'Red', hex = '#FF0000'",
	}
	for _, cmd := range cmds {
		if _, err := testArcadeDB.ExecuteCommand(db, "sql", cmd); err != nil {
			t.Fatalf("setup command %q failed: %v", cmd, err)
		}
	}

	ds := newTestDatasource(t, db)
	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{
			makeQuery("C", QueryModel{
				QueryMode: "gremlin",
				RawQuery:  "g.V().hasLabel('Color')",
			}),
		},
	})
	if err != nil {
		t.Fatalf("QueryData returned top-level error: %v", err)
	}

	r, ok := resp.Responses["C"]
	if !ok {
		t.Fatal("expected response for refID 'C'")
	}
	if r.Error != nil {
		t.Fatalf("expected no error for refID 'C', got: %v", r.Error)
	}
	if len(r.Frames) == 0 {
		t.Fatal("expected at least one frame in Gremlin response")
	}
}

func TestQueryData_RoutesToTimeSeries(t *testing.T) {
	db := uniqueDBName("ds_ts")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	ds := newTestDatasource(t, db)

	// Time series queries hit the /grafana/query endpoint which may not exist
	// on a plain database. We just verify that the routing reaches the time
	// series handler (the response will contain an error from ArcadeDB, not a
	// routing error like "unknown query mode").
	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{
			makeQuery("D", QueryModel{
				QueryMode: "timeseries",
				TSType:    "metric",
				TSFields:  []string{"value"},
			}),
		},
	})
	if err != nil {
		t.Fatalf("QueryData returned top-level error: %v", err)
	}

	r, ok := resp.Responses["D"]
	if !ok {
		t.Fatal("expected response for refID 'D'")
	}

	// The response may have an error (because the DB has no time series data),
	// but the important thing is that routing worked - no "unknown query mode" error.
	if r.Error != nil && r.Error.Error() == "unknown query mode: timeseries" {
		t.Fatal("timeseries query was not routed correctly")
	}
	t.Logf("timeseries routing result: error=%v, frames=%d", r.Error, len(r.Frames))
}

func TestQueryData_MultipleQueries(t *testing.T) {
	db := uniqueDBName("ds_multi")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	cmds := []string{
		"CREATE DOCUMENT TYPE TableA",
		"CREATE DOCUMENT TYPE TableB",
		"INSERT INTO TableA SET x = 1",
		"INSERT INTO TableB SET y = 2",
	}
	for _, cmd := range cmds {
		if _, err := testArcadeDB.ExecuteCommand(db, "sql", cmd); err != nil {
			t.Fatalf("setup command %q failed: %v", cmd, err)
		}
	}

	ds := newTestDatasource(t, db)
	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{
			makeQuery("Q1", QueryModel{QueryMode: "sql", RawQuery: "SELECT x FROM TableA"}),
			makeQuery("Q2", QueryModel{QueryMode: "sql", RawQuery: "SELECT y FROM TableB"}),
		},
	})
	if err != nil {
		t.Fatalf("QueryData returned top-level error: %v", err)
	}

	if _, ok := resp.Responses["Q1"]; !ok {
		t.Error("missing response for refID 'Q1'")
	}
	if _, ok := resp.Responses["Q2"]; !ok {
		t.Error("missing response for refID 'Q2'")
	}

	if resp.Responses["Q1"].Error != nil {
		t.Errorf("Q1 error: %v", resp.Responses["Q1"].Error)
	}
	if resp.Responses["Q2"].Error != nil {
		t.Errorf("Q2 error: %v", resp.Responses["Q2"].Error)
	}

	if len(resp.Responses["Q1"].Frames) == 0 {
		t.Error("Q1 has no frames")
	}
	if len(resp.Responses["Q2"].Frames) == 0 {
		t.Error("Q2 has no frames")
	}
}

func TestQueryData_InvalidMode(t *testing.T) {
	db := uniqueDBName("ds_invalid")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	ds := newTestDatasource(t, db)
	resp, err := ds.QueryData(context.Background(), &backend.QueryDataRequest{
		Queries: []backend.DataQuery{
			makeQuery("X", QueryModel{QueryMode: "nosql", RawQuery: "SELECT 1"}),
		},
	})
	// The error should be per-query, not top-level.
	if err != nil {
		t.Fatalf("expected no top-level error, got: %v", err)
	}

	r, ok := resp.Responses["X"]
	if !ok {
		t.Fatal("expected response for refID 'X'")
	}
	if r.Error == nil {
		t.Fatal("expected per-query error for unknown mode")
	}
	if r.Error.Error() != "unknown query mode: nosql" {
		t.Errorf("unexpected error message: %v", r.Error)
	}
}

func TestCheckHealth_Healthy(t *testing.T) {
	db := uniqueDBName("ds_health_ok")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	ds := newTestDatasource(t, db)
	result, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if err != nil {
		t.Fatalf("CheckHealth returned error: %v", err)
	}
	if result.Status != backend.HealthStatusOk {
		t.Errorf("expected HealthStatusOk, got %v (message: %s)", result.Status, result.Message)
	}
}

func TestCheckHealth_Unhealthy(t *testing.T) {
	// Use bad credentials to trigger a health check failure.
	ds := &Datasource{
		client: NewArcadeDBClient(testArcadeDB.BaseURL(), "nonexistent_db_xyz", "wrong_user", "wrong_pass"),
	}

	result, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	if err != nil {
		t.Fatalf("CheckHealth returned top-level error: %v", err)
	}
	if result.Status != backend.HealthStatusError {
		t.Errorf("expected HealthStatusError, got %v (message: %s)", result.Status, result.Message)
	}
}

func TestCallResource_Metadata(t *testing.T) {
	db := uniqueDBName("ds_meta")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	ds := newTestDatasource(t, db)

	var captured *backend.CallResourceResponse
	sender := callResourceResponseSenderFunc(func(resp *backend.CallResourceResponse) error {
		captured = resp
		return nil
	})

	err := ds.CallResource(context.Background(), &backend.CallResourceRequest{
		Path:   "metadata",
		Method: http.MethodGet,
	}, sender)
	if err != nil {
		t.Fatalf("CallResource returned error: %v", err)
	}
	if captured == nil {
		t.Fatal("sender was never called")
	}

	// The metadata endpoint may return 200 or 500 depending on whether the
	// ArcadeDB instance supports the /grafana/metadata endpoint. Either way,
	// the sender must have been invoked with a response.
	t.Logf("metadata response status: %d, body length: %d", captured.Status, len(captured.Body))
}

func TestCallResource_NotFound(t *testing.T) {
	db := uniqueDBName("ds_notfound")
	if err := testArcadeDB.CreateDatabase(db); err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer testArcadeDB.DropDatabase(db)

	ds := newTestDatasource(t, db)

	var captured *backend.CallResourceResponse
	sender := callResourceResponseSenderFunc(func(resp *backend.CallResourceResponse) error {
		captured = resp
		return nil
	})

	err := ds.CallResource(context.Background(), &backend.CallResourceRequest{
		Path:   "unknown/path",
		Method: http.MethodGet,
	}, sender)
	if err != nil {
		t.Fatalf("CallResource returned error: %v", err)
	}
	if captured == nil {
		t.Fatal("sender was never called")
	}
	if captured.Status != http.StatusNotFound {
		t.Errorf("expected 404, got %d", captured.Status)
	}
}
