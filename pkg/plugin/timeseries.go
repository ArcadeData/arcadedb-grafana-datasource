package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// HandleTimeSeriesQuery processes a Time Series mode query.
func HandleTimeSeriesQuery(ctx context.Context, client *ArcadeDBClient, query backend.DataQuery, qm QueryModel) (data.Frames, error) {
	target := GrafanaQueryTarget{
		RefID:  query.RefID,
		Type:   qm.TSType,
		Fields: qm.TSFields,
		Tags:   qm.TSTags,
	}

	if qm.TSAggregation != nil && len(qm.TSAggregation.Requests) > 0 {
		target.Aggregation = &GrafanaAggregation{
			BucketInterval: qm.TSAggregation.BucketInterval,
			Requests:       qm.TSAggregation.Requests,
		}
	}

	request := &GrafanaQueryRequest{
		From:          query.TimeRange.From.UnixMilli(),
		To:            query.TimeRange.To.UnixMilli(),
		MaxDataPoints: query.MaxDataPoints,
		Targets:       []GrafanaQueryTarget{target},
	}

	results, err := client.QueryTimeSeries(ctx, request)
	if err != nil {
		return nil, err
	}

	// Parse the response for our refID
	rawResult, ok := results[query.RefID]
	if !ok {
		return nil, fmt.Errorf("no result for refID %s", query.RefID)
	}

	return parseDataFrameResponse(rawResult)
}

// parseDataFrameResponse parses ArcadeDB's Grafana DataFrame wire format into Grafana data frames.
func parseDataFrameResponse(raw json.RawMessage) (data.Frames, error) {
	var result struct {
		Frames []struct {
			Schema struct {
				Fields []struct {
					Name string `json:"name"`
					Type string `json:"type"`
				} `json:"fields"`
			} `json:"schema"`
			Data struct {
				Values []json.RawMessage `json:"values"`
			} `json:"data"`
		} `json:"frames"`
	}

	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("failed to parse DataFrame response: %w", err)
	}

	var frames data.Frames
	for _, f := range result.Frames {
		frame := data.NewFrame("")

		for i, field := range f.Schema.Fields {
			if i >= len(f.Data.Values) {
				continue
			}

			switch field.Type {
			case "time":
				var values []float64
				if err := json.Unmarshal(f.Data.Values[i], &values); err != nil {
					return nil, fmt.Errorf("failed to parse time values: %w", err)
				}
				times := make([]time.Time, len(values))
				for j, v := range values {
					times[j] = time.UnixMilli(int64(v))
				}
				frame.Fields = append(frame.Fields, data.NewField(field.Name, nil, times))

			case "number":
				var values []*float64
				if err := json.Unmarshal(f.Data.Values[i], &values); err != nil {
					return nil, fmt.Errorf("failed to parse number values: %w", err)
				}
				frame.Fields = append(frame.Fields, data.NewField(field.Name, nil, values))

			case "string":
				var values []*string
				if err := json.Unmarshal(f.Data.Values[i], &values); err != nil {
					return nil, fmt.Errorf("failed to parse string values: %w", err)
				}
				frame.Fields = append(frame.Fields, data.NewField(field.Name, nil, values))

			case "boolean":
				var values []*bool
				if err := json.Unmarshal(f.Data.Values[i], &values); err != nil {
					return nil, fmt.Errorf("failed to parse boolean values: %w", err)
				}
				frame.Fields = append(frame.Fields, data.NewField(field.Name, nil, values))

			default:
				var values []*string
				if err := json.Unmarshal(f.Data.Values[i], &values); err != nil {
					return nil, fmt.Errorf("failed to parse values for field %s: %w", field.Name, err)
				}
				frame.Fields = append(frame.Fields, data.NewField(field.Name, nil, values))
			}
		}

		frames = append(frames, frame)
	}

	return frames, nil
}
