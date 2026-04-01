package plugin

import (
	"context"
	"encoding/json"
	"fmt"

	"net/http"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
)

type Datasource struct {
	client *ArcadeDBClient
}

type DatasourceSettings struct {
	Database string `json:"database"`
}

var _ backend.QueryDataHandler = (*Datasource)(nil)
var _ backend.CheckHealthHandler = (*Datasource)(nil)
var _ backend.CallResourceHandler = (*Datasource)(nil)

func NewDatasource(_ context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	var jsonData DatasourceSettings
	if err := json.Unmarshal(settings.JSONData, &jsonData); err != nil {
		return nil, fmt.Errorf("failed to parse data source settings: %w", err)
	}

	client := NewArcadeDBClient(settings.URL, jsonData.Database, settings.BasicAuthUser, settings.DecryptedSecureJSONData["password"])

	return &Datasource{client: client}, nil
}

func (d *Datasource) Dispose() {
	// Cleanup if needed
}

func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		var qm QueryModel
		if err := json.Unmarshal(q.JSON, &qm); err != nil {
			response.Responses[q.RefID] = backend.DataResponse{Error: fmt.Errorf("failed to parse query: %w", err)}
			continue
		}

		// Default to SQL if no query mode specified
		if qm.QueryMode == "" {
			qm.QueryMode = "sql"
		}

		var resp backend.DataResponse
		switch qm.QueryMode {
		case "timeseries":
			resp = d.queryTimeSeries(ctx, q, qm)
		case "sql":
			resp = d.queryCommand(ctx, q, qm, "sql")
		case "cypher":
			resp = d.queryCommand(ctx, q, qm, "cypher")
		case "gremlin":
			resp = d.queryCommand(ctx, q, qm, "gremlin")
		default:
			resp = backend.DataResponse{Error: fmt.Errorf("unknown query mode: %s", qm.QueryMode)}
		}

		response.Responses[q.RefID] = resp
	}

	return response, nil
}

func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	err := d.client.CheckHealth(ctx)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Failed to connect to ArcadeDB: %s", err.Error()),
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: fmt.Sprintf("Connected to ArcadeDB database '%s'", d.client.database),
	}, nil
}

func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	if req.Path == "metadata" {
		metadata, err := d.client.GetMetadata(ctx)
		if err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body:   []byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())),
			})
		}

		body, err := json.Marshal(metadata)
		if err != nil {
			return sender.Send(&backend.CallResourceResponse{
				Status: http.StatusInternalServerError,
				Body:   []byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())),
			})
		}

		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusOK,
			Headers: map[string][]string{"Content-Type": {"application/json"}},
			Body:   body,
		})
	}

	return sender.Send(&backend.CallResourceResponse{Status: http.StatusNotFound})
}

// queryTimeSeries handles Time Series mode queries via /grafana/query endpoint.
func (d *Datasource) queryTimeSeries(ctx context.Context, query backend.DataQuery, qm QueryModel) backend.DataResponse {
	frames, err := HandleTimeSeriesQuery(ctx, d.client, query, qm)
	if err != nil {
		return backend.DataResponse{Error: err}
	}
	return backend.DataResponse{Frames: frames}
}

// queryCommand handles SQL/Cypher/Gremlin queries via /command endpoint.
func (d *Datasource) queryCommand(ctx context.Context, query backend.DataQuery, qm QueryModel, language string) backend.DataResponse {
	frames, err := HandleCommandQuery(ctx, d.client, query, qm, language)
	if err != nil {
		return backend.DataResponse{Error: err}
	}
	return backend.DataResponse{Frames: frames}
}

