# CLAUDE.md

Instructions for Claude Code when working on this project.

## Project Overview

This is a Grafana data source plugin for ArcadeDB. It has a TypeScript/React frontend and a Go backend. It connects to ArcadeDB's HTTP REST API to execute queries and return results in Grafana's data frame format.

## Response Formatting
- Never use the em dash character in responses. Use a normal dash, a comma, or rephrase instead.
- Don't add Claude or Claude Code as author anywhere (commits, comments, code, etc.)

## Tech Stack

- **Frontend**: TypeScript, React, Grafana UI components (`@grafana/ui`, `@grafana/data`, `@grafana/runtime`)
- **Backend**: Go, Grafana Plugin SDK (`grafana-plugin-sdk-go`)
- **Build**: npm (frontend), Mage (Go backend)
- **Testing**: Jest (frontend), Go test (backend)
- **License**: Apache 2.0

## Build Commands

### Frontend
- `npm install` - Install dependencies
- `npm run dev` - Watch mode (rebuilds on changes)
- `npm run build` - Production build to `dist/`
- `npm run test` - Run Jest tests
- `npm run lint` - Lint TypeScript
- `npm run typecheck` - Type check without emitting

### Backend
- `mage -v build:linux` - Build Go backend for Linux (amd64 + arm64)
- `mage -v build:darwin` - Build Go backend for macOS (amd64 + arm64)
- `mage -v build:windows` - Build Go backend for Windows (amd64)
- `go test ./pkg/...` - Run Go tests
- `go test -race ./pkg/...` - Run Go tests with race detector

### Development Environment
- `docker compose up -d` - Start Grafana + ArcadeDB dev containers
- Grafana at http://localhost:3000 (admin/admin)
- ArcadeDB at http://localhost:2480 (root/arcadedb)
- Plugin is auto-provisioned via `provisioning/datasources/arcadedb.yaml`

## Project Structure

```
src/                          # Frontend (TypeScript/React)
  plugin.json                 # Plugin metadata (id, capabilities)
  module.ts                   # Entry point - registers the plugin
  datasource.ts               # DataSourceApi - delegates to Go backend
  types.ts                    # All shared TypeScript types
  components/
    ConfigEditor.tsx           # Data source settings form (URL, DB, auth)
    QueryEditor.tsx            # Top-level editor with mode tabs
    TimeSeriesEditor.tsx       # Visual builder for TS queries
    SqlEditor.tsx              # Monaco editor for SQL
    CypherEditor.tsx           # Monaco editor for Cypher
    GremlinEditor.tsx          # Monaco editor for Gremlin
    VariableQueryEditor.tsx    # Template variable query editor

pkg/plugin/                   # Backend (Go)
  main.go                     # Entry point - starts gRPC plugin server
  datasource.go               # Implements QueryDataHandler, CheckHealthHandler
  arcadedb_client.go          # HTTP client wrapping all ArcadeDB REST calls
  timeseries.go               # Translates TS queries to /grafana/query format
  command.go                  # Handles SQL/Cypher/Gremlin via /command endpoint
  nodegraph.go                # Converts ArcadeDB graph JSON to Node Graph frames
  macros.go                   # $__timeFrom, $__timeTo, $__timeFilter, $__interval
  models.go                   # Go struct definitions
```

## ArcadeDB REST API Contracts

The Go backend communicates with these ArcadeDB endpoints:

### Health Check
```
GET /api/v1/ts/{database}/grafana/health
Authorization: Basic <base64>

Response: {"status": "ok", "database": "mydb"}
```

### Metadata Discovery
```
GET /api/v1/ts/{database}/grafana/metadata
Authorization: Basic <base64>

Response:
{
  "types": [
    {
      "name": "weather",
      "fields": [{"name": "temperature", "dataType": "DOUBLE"}],
      "tags": [{"name": "location", "dataType": "STRING"}]
    }
  ],
  "aggregationTypes": ["SUM", "AVG", "MIN", "MAX", "COUNT"]
}
```

### Time Series Query
```
POST /api/v1/ts/{database}/grafana/query
Authorization: Basic <base64>
Content-Type: application/json

Request:
{
  "from": 1000,
  "to": 3000,
  "maxDataPoints": 1000,
  "targets": [
    {
      "refId": "A",
      "type": "weather",
      "fields": ["temperature"],
      "tags": {"location": "us-east"},
      "aggregation": {
        "bucketInterval": 5000,
        "requests": [{"field": "temperature", "type": "AVG", "alias": "avg_temp"}]
      }
    }
  ]
}

Response (Grafana DataFrame wire format):
{
  "results": {
    "A": {
      "frames": [{
        "schema": {
          "fields": [
            {"name": "ts", "type": "time"},
            {"name": "temperature", "type": "number"}
          ]
        },
        "data": {
          "values": [[1000, 2000], [22.5, 18.3]]
        }
      }]
    }
  }
}
```

### Command Execution (SQL / Cypher / Gremlin)
```
POST /api/v1/command/{database}
Authorization: Basic <base64>
Content-Type: application/json

Request:
{
  "language": "sql" | "cypher" | "gremlin",
  "command": "SELECT * FROM Person LIMIT 10",
  "serializer": "record" | "graph",
  "limit": 1000
}

Response with serializer "record":
{
  "user": "root",
  "version": "...",
  "serverName": "...",
  "result": [
    {"@rid": "#12:0", "@type": "Person", "name": "Alice", "age": 30},
    ...
  ]
}

Response with serializer "graph":
{
  "user": "root",
  "result": {
    "vertices": [
      {"r": "#12:0", "t": "Person", "p": {"name": "Alice", "age": 30}, "i": 3, "o": 5}
    ],
    "edges": [
      {"r": "#20:0", "t": "FriendOf", "p": {"since": 2020}, "i": "#12:1", "o": "#12:0"}
    ]
  }
}
```

### Graph Serializer Format (critical for nodegraph.go)

**Vertices**: `r` = RID, `t` = type name, `p` = properties, `i` = incoming edge count (int), `o` = outgoing edge count (int)

**Edges**: `r` = RID, `t` = type name, `p` = properties, `i` = IN vertex RID (string), `o` = OUT vertex RID (string)

### Grafana Node Graph Frame Requirements

**Nodes frame** fields: `id` (string, required), `title`, `subtitle`, `mainstat`, `secondarystat`, `arc__*` (float 0-1), `detail__*`, `color`, `icon`

**Edges frame** fields: `id` (string, required), `source` (required, node id), `target` (required, node id), `mainstat`, `secondarystat`, `detail__*`

Set `frame.Meta.PreferredVisualization = "nodeGraph"` on both frames.

### Mapping: ArcadeDB graph -> Node Graph

- Vertex `r` -> node `id`
- First of vertex `p.name`, `p.label`, `p.title`, or `r` -> node `title`
- Vertex `t` -> node `subtitle`
- All vertex `p.*` -> node `detail__*`
- Edge `r` -> edge `id`
- Edge `o` (OUT vertex) -> edge `source`
- Edge `i` (IN vertex) -> edge `target`
- Edge `t` -> edge `mainstat`
- All edge `p.*` -> edge `detail__*`

## Coding Conventions

- Frontend: follow Grafana plugin conventions (functional React components, hooks)
- Backend: standard Go conventions (gofmt, go vet)
- Keep dependencies minimal
- All new code must have tests
- Error messages should be user-friendly (shown in Grafana UI)

## Macro Expansion (macros.go)

Applied to raw query strings before sending to ArcadeDB:

- `$__timeFrom` -> epoch milliseconds of time range start
- `$__timeTo` -> epoch milliseconds of time range end
- `$__timeFilter(column)` -> `column >= {from} AND column <= {to}`
- `$__interval` -> auto-calculated step as duration string (e.g., `60s`, `5m`)

## Testing Strategy

### Go Backend Tests (testcontainers)

Tests run against a real ArcadeDB instance managed by testcontainers. A single container is started via `TestMain` in `testutil_test.go` and shared across all test functions in the package. Docker must be running.

- Each test creates its own database with `uniqueDBName()` and defers cleanup with `DropDatabase()` for isolation.
- Use `testArcadeDB.NewTestClient(db)` to get a client pointed at the test instance.
- Use `testArcadeDB.ExecuteCommand(db, language, command)` to set up test data.
- Do not use `httptest.NewServer` or mock HTTP responses. All backend tests hit the real ArcadeDB API.
- Test files live alongside their source: `pkg/plugin/*_test.go`.

### Frontend Tests (Jest + React Testing Library)

Tests run in jsdom with mocked Grafana runtime dependencies. No running backend needed.

- Mock `@grafana/ui` components as simple HTML elements (see `QueryEditor.test.tsx` for the pattern). Use `data-testid` attributes for querying.
- Mock `@grafana/runtime` to stub `DataSourceWithBackend` and `getTemplateSrv` (see `datasource.test.ts` for the pattern).
- Mock child editor components as stubs when testing the parent `QueryEditor` (e.g., `jest.mock('./SqlEditor', ...)`).
- Use a `makeProps()` or `createDataSource()` helper to build test fixtures with sensible defaults and allow overrides.
- Use `jest.fn()` for `onChange` and `onRunQuery` callbacks, then assert they were called with expected arguments.
- Test files live next to their source: `src/**/*.test.ts(x)`.
- Run `npm run test:ci` (not `npm run test`) for single-run execution in CI or before committing.
