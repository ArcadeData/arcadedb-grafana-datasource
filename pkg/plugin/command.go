package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// HandleCommandQuery processes SQL/Cypher/Gremlin queries via the /command endpoint.
func HandleCommandQuery(ctx context.Context, client *ArcadeDBClient, query backend.DataQuery, qm QueryModel, language string) (data.Frames, error) {
	// Expand macros in the raw query
	interval := query.Interval
	if interval == 0 {
		interval = time.Second * 60
	}
	expandedQuery := ExpandMacros(qm.RawQuery, query.TimeRange, interval)

	// Choose serializer based on node graph mode
	serializer := "record"
	if qm.NodeGraphEnabled && (language == "cypher" || language == "gremlin") {
		serializer = "graph"
	}

	req := &CommandRequest{
		Language:   language,
		Command:    expandedQuery,
		Serializer: serializer,
		Limit:      1000,
	}

	respBody, err := client.ExecuteCommand(ctx, req)
	if err != nil {
		return nil, err
	}

	if serializer == "graph" {
		return parseGraphResponse(respBody)
	}

	return parseRecordResponse(respBody)
}

// parseRecordResponse converts a flat JSON array result into Grafana data frames.
func parseRecordResponse(body []byte) (data.Frames, error) {
	var resp CommandResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse command response: %w", err)
	}

	if len(resp.Result) == 0 {
		frame := data.NewFrame("result")
		return data.Frames{frame}, nil
	}

	// Collect all unique column names in order
	columnOrder := []string{}
	columnSet := map[string]bool{}
	for _, row := range resp.Result {
		for key := range row {
			if key == "@rid" || key == "@type" || key == "@cat" {
				continue
			}
			if !columnSet[key] {
				columnSet[key] = true
				columnOrder = append(columnOrder, key)
			}
		}
	}

	// Detect column types from first non-nil value
	frame := data.NewFrame("result")
	for _, col := range columnOrder {
		field := buildField(col, resp.Result)
		if field != nil {
			frame.Fields = append(frame.Fields, field)
		}
	}

	return data.Frames{frame}, nil
}

// buildField creates a data.Field from a column name and result rows.
func buildField(name string, rows []map[string]interface{}) *data.Field {
	// Detect type from first non-nil value
	var fieldType string
	for _, row := range rows {
		v, ok := row[name]
		if !ok || v == nil {
			continue
		}
		switch v.(type) {
		case float64:
			fieldType = "number"
		case bool:
			fieldType = "boolean"
		default:
			fieldType = "string"
		}
		break
	}

	if fieldType == "" {
		fieldType = "string"
	}

	switch fieldType {
	case "number":
		values := make([]*float64, len(rows))
		for i, row := range rows {
			if v, ok := row[name]; ok && v != nil {
				f := v.(float64)
				values[i] = &f
			}
		}
		return data.NewField(name, nil, values)

	case "boolean":
		values := make([]*bool, len(rows))
		for i, row := range rows {
			if v, ok := row[name]; ok && v != nil {
				b := v.(bool)
				values[i] = &b
			}
		}
		return data.NewField(name, nil, values)

	default:
		values := make([]*string, len(rows))
		for i, row := range rows {
			if v, ok := row[name]; ok && v != nil {
				s := fmt.Sprintf("%v", v)
				values[i] = &s
			}
		}
		return data.NewField(name, nil, values)
	}
}

// parseGraphResponse converts a graph serializer response into Node Graph frames.
func parseGraphResponse(body []byte) (data.Frames, error) {
	var resp GraphResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse graph response: %w", err)
	}

	return BuildNodeGraphFrames(resp.Result.Vertices, resp.Result.Edges)
}
