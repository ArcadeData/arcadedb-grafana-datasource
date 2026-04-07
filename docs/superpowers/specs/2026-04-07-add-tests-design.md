# Test Suite Design - ArcadeDB Grafana Data Source Plugin

## Overview

Add comprehensive test coverage across the Go backend and TypeScript frontend, plus a GitHub Actions CI pipeline to gate PRs. Go tests use [testcontainers-go](https://golang.testcontainers.org/) to run tests against real ArcadeDB instances. Frontend tests use React Testing Library with mocked Grafana runtime APIs.

## Approach

Bottom-up: start with the lowest-level Go code (HTTP client), work up through parsers and routing, then frontend components, then CI. Each layer validates the foundation the next depends on.

## Go Backend

### Shared Testcontainers Helper

**File**: `pkg/plugin/testutil_test.go`

A `TestArcadeDB` struct wrapping a testcontainers ArcadeDB instance:

- Starts `arcadedata/arcadedb:latest`, waits for port 2480 readiness
- Helper methods: `CreateDatabase(name)`, `ExecuteCommand(db, language, command)`, `BaseURL()`, `Cleanup()`
- Uses `TestMain(m *testing.M)` to start one container per package test run (not per test)
- Each test creates its own database with a unique name to avoid cross-test contamination

### arcadedb_client_test.go

Tests for the core HTTP client against a real ArcadeDB:

- `TestCheckHealth` - verify healthy response for existing DB; error for nonexistent DB
- `TestGetMetadata` - create DB with document type + fields, verify metadata response
- `TestQueryTimeSeries` - create DB with time series type, insert records, query and verify
- `TestExecuteCommand_SQL` - create DB with document type, insert records, execute SQL SELECT
- `TestExecuteCommand_OpenCypher` - same pattern with openCypher query
- `TestExecuteCommand_GraphSerializer` - create vertices + edges, execute with `serializer: "graph"`, verify structure
- `TestExecuteCommand_ErrorHandling` - invalid SQL, wrong credentials, nonexistent database
- `TestDoRequest_Timeout` - verify client respects timeout settings

### command_test.go

Tests for SQL/openCypher/Gremlin query execution and response parsing:

- `TestHandleCommandQuery_SQL` - insert mixed-type records (strings, ints, floats, booleans, nulls), verify Grafana data frame field types and values
- `TestHandleCommandQuery_OpenCypher` - same pattern with openCypher syntax
- `TestHandleCommandQuery_Gremlin` - same with Gremlin
- `TestHandleCommandQuery_GraphMode` - create vertices + edges, query with `nodeGraph: true`, verify Node Graph frames returned
- `TestHandleCommandQuery_EmptyResult` - query returning no rows, verify empty frame without error
- `TestHandleCommandQuery_NullValues` - records with null/missing fields, verify graceful handling
- `TestParseRecordResponse_FieldTypeDetection` - insert records with DOUBLE, INTEGER, STRING, BOOLEAN, DATE fields, verify `buildField` infers correct Grafana types

### timeseries_test.go

Tests for time series response parsing:

- `TestHandleTimeSeriesQuery_Basic` - create time series type, insert timestamped records, query a time range, verify frame
- `TestHandleTimeSeriesQuery_WithAggregation` - query with aggregation bucket, verify aggregated results
- `TestHandleTimeSeriesQuery_WithTags` - filter by tag, verify only matching records
- `TestHandleTimeSeriesQuery_EmptyRange` - query time range with no data, verify empty frame
- `TestHandleTimeSeriesQuery_MultipleFields` - select multiple fields, verify all appear in frame

### nodegraph_test.go (extend existing)

Additional tests using real graph data from ArcadeDB:

- `TestBuildNodeGraphFrames_RealGraphData` - create vertices + edges, fetch via graph serializer, verify node IDs/titles/subtitles and edge source/target mapping
- `TestBuildNodeGraphFrames_VertexProperties` - verify all vertex properties appear as `detail__*` fields
- `TestBuildNodeGraphFrames_EdgeProperties` - verify edge properties appear as `detail__*` fields
- `TestBuildNodeGraphFrames_TitleResolution` - create vertices with `name`, `label`, `title`, and none, verify fallback chain
- `TestBuildNodeGraphFrames_NullProperties` - vertices/edges with missing optional properties

### macros_test.go (extend existing)

Additional macro expansion tests:

- `TestExpandMacros_MultipleMacrosInSameQuery` - query containing all four macros at once
- `TestExpandMacros_NoMacros` - plain query passes through unchanged
- `TestExpandMacros_RepeatedMacro` - same macro appearing twice in one query

### datasource_test.go

Integration tests for the Grafana plugin entry point:

- `TestQueryData_RoutesToTimeSeries` - query with `queryMode: "timeseries"`, verify time series handler reached
- `TestQueryData_RoutesToSQL` - query with `queryMode: "sql"`, verify command handler reached
- `TestQueryData_RoutesToOpenCypher` - same for openCypher
- `TestQueryData_RoutesToGremlin` - same for Gremlin
- `TestQueryData_MultipleQueries` - multiple queries with different modes in one request, verify each routed correctly
- `TestQueryData_InvalidMode` - unknown query mode, verify meaningful error
- `TestCheckHealth_Healthy` - valid credentials, verify health check passes
- `TestCheckHealth_Unhealthy` - wrong credentials or unreachable host, verify error
- `TestCallResource_Metadata` - call `/metadata` resource endpoint, verify type/field metadata

## TypeScript Frontend

All frontend tests use React Testing Library with `@testing-library/user-event`. Grafana's `@grafana/runtime` is mocked at the module level (specifically `getBackendSrv` and `getTemplateSrv`).

### src/components/ConfigEditor.test.tsx

- `renders all fields` - verify URL, database, username, password inputs present
- `calls onChange when URL is updated` - type in URL, verify `onOptionsChange` fires
- `calls onChange when database is updated` - same for database
- `calls onChange when username is updated` - same for username
- `handles password set and reset` - verify SecretInput state, click reset, verify `secureJsonFields` cleared
- `populates fields from existing options` - pass pre-filled props, verify inputs show saved values

### src/components/QueryEditor.test.tsx

- `renders with default SQL mode` - verify SQL editor shown
- `switches to OpenCypher mode` - select openCypher from dropdown, verify editor renders
- `switches to Gremlin mode` - same for Gremlin
- `switches to TimeSeries mode` - same for TimeSeries editor
- `shows node graph toggle for command modes` - verify toggle visible in SQL/openCypher/Gremlin
- `hides node graph toggle for timeseries mode` - verify toggle hidden

### src/components/TimeSeriesEditor.test.tsx

- `renders loading state while fetching metadata` - mock slow response, verify loading indicator
- `renders type selector after metadata loads` - mock metadata with types, verify dropdown populated
- `selecting a type filters available fields` - select type, verify field checkboxes match
- `selecting fields updates query` - check fields, verify `onChange` fires with updated array
- `renders tag inputs` - select type with tags, verify tag filter inputs
- `renders aggregation options` - verify aggregation dropdown and bucket interval input

### src/components/SqlEditor.test.tsx, CypherEditor.test.tsx, GremlinEditor.test.tsx

- `renders code editor with current query` - verify editor shows query text
- `calls onChange on blur` - simulate blur, verify onChange fires

### src/components/VariableQueryEditor.test.tsx

- `renders with query text` - verify editor renders with existing variable query
- `calls onChange when query changes` - update query, verify onChange

### src/datasource.test.ts

- `filterQuery excludes empty queries` - verify returns false for blank strings
- `filterQuery includes non-empty queries` - verify returns true
- `applyTemplateVariables replaces variables` - mock `getTemplateSrv().replace()`, verify interpolation
- `getDefaultQuery returns expected defaults` - verify default query mode and structure

## GitHub Actions CI

**File**: `.github/workflows/ci.yml`

Triggers on pull requests and pushes to `main`.

### frontend job (ubuntu-latest)

1. `npm ci`
2. `npm run typecheck`
3. `npm run lint`
4. `npm run test:ci` (with `--passWithNoTests` removed so zero tests fails)

### backend job (ubuntu-latest)

1. Set up Go (version from `go.mod`)
2. `go vet ./pkg/...`
3. `go test -race -v ./pkg/...`

The two jobs run in parallel. Testcontainers pulls ArcadeDB automatically - GitHub Actions Ubuntu runners support Docker natively, so no `services:` block or docker compose needed.

## Test File Summary

| File | New/Extend | Approx Tests |
|------|-----------|-------------|
| `pkg/plugin/testutil_test.go` | New | (helper) |
| `pkg/plugin/arcadedb_client_test.go` | New | 8 |
| `pkg/plugin/command_test.go` | New | 7 |
| `pkg/plugin/timeseries_test.go` | New | 5 |
| `pkg/plugin/nodegraph_test.go` | Extend | +5 |
| `pkg/plugin/macros_test.go` | Extend | +3 |
| `pkg/plugin/datasource_test.go` | New | 9 |
| `src/components/ConfigEditor.test.tsx` | New | 6 |
| `src/components/QueryEditor.test.tsx` | New | 6 |
| `src/components/TimeSeriesEditor.test.tsx` | New | 6 |
| `src/components/SqlEditor.test.tsx` | New | 2 |
| `src/components/CypherEditor.test.tsx` | New | 2 |
| `src/components/GremlinEditor.test.tsx` | New | 2 |
| `src/components/VariableQueryEditor.test.tsx` | New | 2 |
| `src/datasource.test.ts` | New | 4 |
| `.github/workflows/ci.yml` | New | - |
| **Total** | | ~67 |

## Dependencies to Add

### Go
- `github.com/testcontainers/testcontainers-go`

### npm (devDependencies)
- `@testing-library/react` (if not already present)
- `@testing-library/user-event`
