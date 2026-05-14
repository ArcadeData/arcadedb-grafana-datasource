# Changelog

## 1.0.0-beta.3

- Attach the SLSA provenance bundle (`*.intoto.jsonl`) as a release asset so the plugin validator can verify the build without GitHub API access. Addresses the catalog's no-provenance-attestation recommendation.

## 1.0.0-beta.2

Resubmission for catalog review.

- Fix `package.json` / `plugin.json` version mismatch flagged by the plugin validator. `package.json` is now the single source of truth for the plugin version.
- Add SLSA build provenance attestation to the release workflow.

## 1.0.0-beta.1

First public release of the ArcadeDB data source for Grafana.

### Features

- Time Series query mode with visual builder, auto-discovered types/fields/tags, and built-in aggregations (SUM, AVG, MIN, MAX, COUNT).
- SQL query mode with macro expansion (`$__timeFrom`, `$__timeTo`, `$__timeFilter`, `$__interval`).
- Cypher query mode with optional Node Graph visualization.
- Gremlin query mode with optional Node Graph visualization.
- Template variable support across all query modes.
- Grafana alerting support via the Go backend.

### Requirements

- Grafana 12.3.0 or later
- ArcadeDB 24.x or later with the HTTP API enabled

### Installation

Install via the Grafana Plugin Catalog (`grafana-cli plugins install arcadedb-arcadedb-datasource`) or download the signed zip from the GitHub release and extract it into your Grafana plugins directory.
