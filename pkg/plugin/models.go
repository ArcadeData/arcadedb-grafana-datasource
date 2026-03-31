package plugin

// QueryModel represents the query sent from the frontend.
type QueryModel struct {
	QueryMode        string            `json:"queryMode"`
	RawQuery         string            `json:"rawQuery"`
	TSType           string            `json:"tsType,omitempty"`
	TSFields         []string          `json:"tsFields,omitempty"`
	TSTags           map[string]string `json:"tsTags,omitempty"`
	TSAggregation    *TSAggregation    `json:"tsAggregation,omitempty"`
	NodeGraphEnabled bool              `json:"nodeGraphEnabled,omitempty"`
}

// TSAggregation represents time series aggregation configuration.
type TSAggregation struct {
	BucketInterval int                    `json:"bucketInterval,omitempty"`
	Requests       []TSAggregationRequest `json:"requests,omitempty"`
}

// TSAggregationRequest represents a single aggregation request.
type TSAggregationRequest struct {
	Field string `json:"field"`
	Type  string `json:"type"`
	Alias string `json:"alias,omitempty"`
}

// ArcadeDB API response structures

// GrafanaQueryRequest is the body sent to POST /api/v1/ts/{db}/grafana/query.
type GrafanaQueryRequest struct {
	From          int64                 `json:"from"`
	To            int64                 `json:"to"`
	MaxDataPoints int64                 `json:"maxDataPoints,omitempty"`
	Targets       []GrafanaQueryTarget  `json:"targets"`
}

// GrafanaQueryTarget is a single target in a Grafana query.
type GrafanaQueryTarget struct {
	RefID       string                 `json:"refId"`
	Type        string                 `json:"type"`
	Fields      []string               `json:"fields,omitempty"`
	Tags        map[string]string      `json:"tags,omitempty"`
	Aggregation *GrafanaAggregation    `json:"aggregation,omitempty"`
}

// GrafanaAggregation is the aggregation block in a target.
type GrafanaAggregation struct {
	BucketInterval int                    `json:"bucketInterval,omitempty"`
	Requests       []TSAggregationRequest `json:"requests,omitempty"`
}

// CommandRequest is the body sent to POST /api/v1/command/{db}.
type CommandRequest struct {
	Language   string `json:"language"`
	Command    string `json:"command"`
	Serializer string `json:"serializer,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

// CommandResponse is the response from /command endpoint with "record" serializer.
type CommandResponse struct {
	Result []map[string]interface{} `json:"result"`
}

// GraphResponse is the response from /command endpoint with "graph" serializer.
type GraphResponse struct {
	Result GraphResult `json:"result"`
}

// GraphResult contains separated vertices and edges.
type GraphResult struct {
	Vertices []GraphElement `json:"vertices"`
	Edges    []GraphElement `json:"edges"`
}

// GraphElement represents a vertex or edge in the graph serializer format.
// For vertices: r=RID, t=type, p=properties, i=inEdgeCount(int), o=outEdgeCount(int)
// For edges: r=RID, t=type, p=properties, i=inVertexRID(string), o=outVertexRID(string)
type GraphElement struct {
	R string                 `json:"r"`
	T string                 `json:"t"`
	P map[string]interface{} `json:"p"`
	I interface{}            `json:"i"` // int for vertices, string for edges
	O interface{}            `json:"o"` // int for vertices, string for edges
}

// MetadataResponse is the response from /grafana/metadata.
type MetadataResponse struct {
	Types            []TSTypeMetadata `json:"types"`
	AggregationTypes []string         `json:"aggregationTypes"`
}

// TSTypeMetadata describes a time series type.
type TSTypeMetadata struct {
	Name   string       `json:"name"`
	Fields []TSColumn   `json:"fields"`
	Tags   []TSColumn   `json:"tags"`
}

// TSColumn describes a time series column.
type TSColumn struct {
	Name     string `json:"name"`
	DataType string `json:"dataType"`
}
