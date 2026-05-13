# ArcadeDB Data Source for Grafana

A Grafana data source plugin for [ArcadeDB](https://arcadedb.com), the multi-model database. Query your graph, time series, document, and vector data directly from Grafana dashboards.

## Features

### Multi-Language Query Support

- **Time Series** - Visual query builder with auto-discovered types, fields, tags, and aggregation functions (SUM, AVG, MIN, MAX, COUNT). No query language required.
- **SQL** - Full ArcadeDB SQL support with syntax highlighting and macro expansion.
- **Cypher** - OpenCypher queries with optional graph visualization via Grafana's Node Graph panel.
- **Gremlin** - Apache TinkerPop Gremlin traversals with optional graph visualization.

### Graph Visualization

Cypher and Gremlin query results can be rendered as interactive network graphs using Grafana's built-in Node Graph panel. Vertices become nodes, edges become connections. Properties are available in detail views.

### Time Series Native Support

Connects directly to ArcadeDB's time series engine with:
- Auto-discovery of time series types, fields, and tags
- Built-in aggregation (SUM, AVG, MIN, MAX, COUNT) with configurable bucket intervals
- Tag-based filtering
- Automatic bucket interval calculation from Grafana's `maxDataPoints`

### Alerting

Full Grafana alerting support via the Go backend component. Create alert rules on any query mode.

### Template Variables

Dashboard variables populated from ArcadeDB queries. Use variables in any query mode for dynamic, interactive dashboards.

### Macros

Convenient macros for time-range interpolation in SQL, Cypher, and Gremlin queries:

| Macro | Description | Example Output |
|---|---|---|
| `$__timeFrom` | Dashboard time range start (epoch ms) | `1711900800000` |
| `$__timeTo` | Dashboard time range end (epoch ms) | `1711987200000` |
| `$__timeFilter(col)` | Time range filter expression | `col >= 1711900800000 AND col <= 1711987200000` |
| `$__interval` | Auto-calculated interval duration | `60s` |

## Requirements

- Grafana 12.3.0 or later
- ArcadeDB 24.x or later with HTTP API enabled

## Installation

### From Grafana Plugin Directory (recommended)

```bash
grafana-cli plugins install arcadedb-arcadedb-datasource
```

Then restart Grafana.

### Manual Installation

1. Download the latest release from the [Releases](https://github.com/ArcadeData/arcadedb-grafana-datasource/releases) page.
2. Extract the archive into your Grafana plugins directory (default: `/var/lib/grafana/plugins/`).
3. Restart Grafana.

For unsigned plugin installations on self-hosted Grafana, add to `grafana.ini`:

```ini
[plugins]
allow_loading_unsigned_plugins = arcadedb-arcadedb-datasource
```

### From Source

```bash
git clone https://github.com/ArcadeData/arcadedb-grafana-datasource.git
cd arcadedb-grafana-datasource
npm install
npm run build
mage -v build:linux    # or build:darwin, build:windows
```

Copy the `dist/` directory to your Grafana plugins directory.

## Configuration

1. In Grafana, go to **Connections > Data Sources > Add data source**.
2. Search for **ArcadeDB** and select it.
3. Configure the connection:

| Field | Description | Example |
|---|---|---|
| **URL** | ArcadeDB server HTTP URL | `http://localhost:2480` |
| **Database** | Database name | `mydb` |
| **Username** | ArcadeDB user | `root` |
| **Password** | ArcadeDB password | `arcadedb` |

4. Click **Save & Test** to verify the connection.

## Usage

### Time Series Mode

The visual query builder auto-discovers your time series types from ArcadeDB's metadata endpoint.

1. Select a **Type** (e.g., `cpu_metrics`, `temperature`).
2. Choose one or more **Fields** to plot.
3. Optionally add **Tag Filters** (e.g., `host = server1`).
4. Optionally configure **Aggregation**: select a function (AVG, SUM, MIN, MAX, COUNT), a field, and a bucket interval.

If no bucket interval is specified, it is auto-calculated from Grafana's `maxDataPoints` setting.

### SQL Mode

Write ArcadeDB SQL queries directly. Use macros for time-range interpolation.

**Table output:**
```sql
SELECT name, age, city FROM Person LIMIT 100
```

**Time series output** (include a timestamp column):
```sql
SELECT ts, temperature FROM weather
WHERE $__timeFilter(ts)
ORDER BY ts
```

**Graph traversal via SQL:**
```sql
SELECT name, out('FriendOf').size() AS friends
FROM Person
WHERE out('FriendOf').size() > 3
```

### Cypher Mode

Write OpenCypher queries. Enable the **Node Graph** toggle to visualize results as an interactive graph.

**Table output:**
```cypher
MATCH (p:Person)-[:FRIEND_OF]->(f:Person)
RETURN p.name AS person, f.name AS friend
LIMIT 50
```

**Graph output** (with Node Graph toggle enabled):
```cypher
MATCH (p:Person)-[r:FRIEND_OF]->(f:Person)
RETURN p, r, f
LIMIT 100
```

### Gremlin Mode

Write Apache TinkerPop Gremlin traversals. The **Node Graph** toggle works the same as in Cypher mode.

**Table output:**
```groovy
g.V().hasLabel('Person').project('name','age').by('name').by('age').limit(50)
```

**Graph output** (with Node Graph toggle enabled):
```groovy
g.V().hasLabel('Person').outE('FriendOf').inV().path().limit(50)
```

### Template Variables

Create dashboard variables backed by ArcadeDB queries:

1. Go to **Dashboard Settings > Variables > New variable**.
2. Set **Type** to **Query** and select your ArcadeDB data source.
3. Enter a query that returns a single column:
   ```sql
   SELECT DISTINCT location FROM weather
   ```
4. Use the variable in panels: `SELECT * FROM weather WHERE location = '$location'`

## Node Graph Details

When the Node Graph toggle is enabled, the Go backend translates ArcadeDB's graph response into Grafana's Node Graph format:

**Nodes** (from vertices):
- `id` - Record ID (e.g., `#12:0`)
- `title` - First available: `name`, `label`, `title` property, or the RID
- `subtitle` - Type name
- `detail__*` - All vertex properties

**Edges** (from edges):
- `id` - Record ID
- `source` - Source vertex RID
- `target` - Target vertex RID
- `mainstat` - Edge type name
- `detail__*` - All edge properties

## Architecture

```
Browser <-> Grafana Frontend (React/TypeScript) <-> Grafana Backend (Go, gRPC) <-> ArcadeDB HTTP API
```

The Go backend communicates with ArcadeDB via its REST API:

| ArcadeDB Endpoint | Used For |
|---|---|
| `GET /api/v1/ts/{db}/grafana/health` | Health check (Save & Test) |
| `GET /api/v1/ts/{db}/grafana/metadata` | Time series type/field/tag discovery |
| `POST /api/v1/ts/{db}/grafana/query` | Time series DataFrame queries |
| `POST /api/v1/command/{db}` | SQL, Cypher, Gremlin commands |

## Development

### Prerequisites

- Node.js 22+
- Go 1.25+
- [Mage](https://magefile.org/) (Go build tool)
- Docker and Docker Compose (for the dev environment)

### Quick Start

```bash
# Start dev environment (Grafana + ArcadeDB with sample data)
docker compose up -d

# Install frontend dependencies
npm install

# Build frontend (watch mode)
npm run dev

# Build Go backend
mage -v build:linux   # or build:darwin

# Run tests
npm run test          # Frontend tests (watch mode)
go test ./pkg/...     # Backend tests
```

Grafana is available at http://localhost:3000 (admin/admin) with the plugin auto-configured.

### Running Tests

#### Frontend (Jest + React Testing Library)

```bash
# Watch mode - reruns tests on file changes (useful during development)
npm run test

# Single run - all tests, no watch (used in CI)
npm run test:ci

# Run a specific test file
npx jest src/components/QueryEditor.test.tsx

# Run tests matching a pattern
npx jest --testPathPattern="TimeSeries"
```

Frontend tests cover:
- Component rendering and user interactions (`src/components/*.test.tsx`)
- DataSource class behavior (`src/datasource.test.ts`)

Tests use mocked Grafana runtime dependencies. No running Grafana or ArcadeDB instance is needed.

#### Backend (Go)

```bash
# Run all backend tests
go test ./pkg/...

# With race detector (matches CI)
go test -race -v ./pkg/... -timeout 300s

# Run a specific test file or function
go test -run TestMacroExpansion ./pkg/plugin/

# With verbose output
go test -v ./pkg/plugin/
```

Backend tests cover:
- ArcadeDB HTTP client (`arcadedb_client_test.go`)
- Time series query building (`timeseries_test.go`)
- SQL/Cypher/Gremlin command handling (`command_test.go`)
- Node Graph frame conversion (`nodegraph_test.go`)
- Macro expansion (`macros_test.go`)
- DataSource health check and query routing (`datasource_test.go`)

Tests use testcontainers to run a real ArcadeDB instance. Docker must be running.

#### Linting and Type Checking

```bash
# TypeScript type checking
npm run typecheck

# ESLint
npm run lint

# Auto-fix lint and formatting issues
npm run lint:fix

# Go vet
go vet ./pkg/...
```

#### CI

The CI pipeline (`.github/workflows/ci.yml`) runs on every PR and push to `main`. It executes:

- **Frontend job**: `npm ci`, typecheck, lint, `npm run test:ci`
- **Backend job**: `go vet`, `go test -race`

### Project Structure

```
src/                         # TypeScript/React frontend
  plugin.json                # Plugin metadata
  module.ts                  # Entry point
  datasource.ts              # DataSourceApi implementation
  types.ts                   # Shared types
  components/
    ConfigEditor.tsx          # Data source configuration
    QueryEditor.tsx           # Query mode router
    TimeSeriesEditor.tsx      # Visual time series builder
    SqlEditor.tsx             # SQL code editor
    CypherEditor.tsx          # Cypher code editor
    GremlinEditor.tsx         # Gremlin code editor
    VariableQueryEditor.tsx   # Template variable editor
pkg/plugin/                  # Go backend
  main.go                    # Entry point
  datasource.go              # QueryDataHandler + CheckHealthHandler
  arcadedb_client.go         # ArcadeDB HTTP client
  timeseries.go              # Time series query handler
  command.go                 # SQL/Cypher/Gremlin handler
  nodegraph.go               # Graph-to-NodeGraph conversion
  macros.go                  # Macro expansion
  models.go                  # Shared structs
provisioning/                # Grafana provisioning for dev
docker-compose.yaml          # Dev environment
```

## License

Apache License 2.0. See [LICENSE](LICENSE) for details.
