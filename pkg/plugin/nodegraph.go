package plugin

import (
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// BuildNodeGraphFrames converts ArcadeDB graph elements into Grafana Node Graph frames.
func BuildNodeGraphFrames(vertices []GraphElement, edges []GraphElement) (data.Frames, error) {
	nodesFrame := buildNodesFrame(vertices)
	edgesFrame := buildEdgesFrame(edges)

	// Set preferred visualization
	nodesFrame.Meta = &data.FrameMeta{PreferredVisualization: "nodeGraph"}
	edgesFrame.Meta = &data.FrameMeta{PreferredVisualization: "nodeGraph"}

	return data.Frames{nodesFrame, edgesFrame}, nil
}

// buildNodesFrame creates the nodes frame from vertices.
func buildNodesFrame(vertices []GraphElement) *data.Frame {
	ids := make([]*string, len(vertices))
	titles := make([]*string, len(vertices))
	subtitles := make([]*string, len(vertices))
	mainstats := make([]*string, len(vertices))

	// Collect all property keys for detail__ fields
	propKeys := collectPropertyKeys(vertices)
	detailFields := make(map[string][]*string)
	for _, key := range propKeys {
		detailFields[key] = make([]*string, len(vertices))
	}

	for i, v := range vertices {
		rid := v.R
		ids[i] = &rid

		title := resolveTitle(v)
		titles[i] = &title

		typeName := v.T
		subtitles[i] = &typeName

		// mainstat: show incoming/outgoing edge counts if available
		stat := fmt.Sprintf("in: %v, out: %v", v.I, v.O)
		mainstats[i] = &stat

		// detail fields from properties
		for _, key := range propKeys {
			if val, ok := v.P[key]; ok && val != nil {
				s := fmt.Sprintf("%v", val)
				detailFields[key][i] = &s
			}
		}
	}

	frame := data.NewFrame("nodes",
		data.NewField("id", nil, ids),
		data.NewField("title", nil, titles),
		data.NewField("subtitle", nil, subtitles),
		data.NewField("mainstat", nil, mainstats),
	)

	for _, key := range propKeys {
		frame.Fields = append(frame.Fields, data.NewField("detail__"+key, nil, detailFields[key]))
	}

	return frame
}

// buildEdgesFrame creates the edges frame from edges.
func buildEdgesFrame(edges []GraphElement) *data.Frame {
	ids := make([]*string, len(edges))
	sources := make([]*string, len(edges))
	targets := make([]*string, len(edges))
	mainstats := make([]*string, len(edges))

	// Collect property keys for detail__ fields
	propKeys := collectPropertyKeys(edges)
	detailFields := make(map[string][]*string)
	for _, key := range propKeys {
		detailFields[key] = make([]*string, len(edges))
	}

	for i, e := range edges {
		rid := e.R
		ids[i] = &rid

		// Edge o = OUT vertex (source), i = IN vertex (target)
		source := fmt.Sprintf("%v", e.O)
		target := fmt.Sprintf("%v", e.I)
		sources[i] = &source
		targets[i] = &target

		typeName := e.T
		mainstats[i] = &typeName

		// detail fields from properties
		for _, key := range propKeys {
			if val, ok := e.P[key]; ok && val != nil {
				s := fmt.Sprintf("%v", val)
				detailFields[key][i] = &s
			}
		}
	}

	frame := data.NewFrame("edges",
		data.NewField("id", nil, ids),
		data.NewField("source", nil, sources),
		data.NewField("target", nil, targets),
		data.NewField("mainstat", nil, mainstats),
	)

	for _, key := range propKeys {
		frame.Fields = append(frame.Fields, data.NewField("detail__"+key, nil, detailFields[key]))
	}

	return frame
}

// resolveTitle picks the best title for a vertex node.
func resolveTitle(v GraphElement) string {
	for _, key := range []string{"name", "label", "title"} {
		if val, ok := v.P[key]; ok && val != nil {
			return fmt.Sprintf("%v", val)
		}
	}
	return v.R
}

// collectPropertyKeys returns all unique property keys in order of first appearance.
func collectPropertyKeys(elements []GraphElement) []string {
	seen := map[string]bool{}
	var keys []string
	for _, e := range elements {
		for key := range e.P {
			if !seen[key] {
				seen[key] = true
				keys = append(keys, key)
			}
		}
	}
	return keys
}
