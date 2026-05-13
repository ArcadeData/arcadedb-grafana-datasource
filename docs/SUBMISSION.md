# Grafana Plugin Catalog Submission

Internal notes for shipping `arcadedb-arcadedb-datasource` to the catalog. Not bundled in the release zip.

## Identifiers

- Plugin ID: `arcadedb-arcadedb-datasource`
- Org account on grafana.com: Arcade Data Ltd
- Submission URL: https://grafana.com/auth/sign-in -> My Plugins -> Submit New Plugin
- Signing token secret: `GRAFANA_ACCESS_POLICY_TOKEN` (scope `plugins:write`)
- Primary contact: r.franchini@arcadedata.com

## Reviewer setup

The plugin ships with a `docker-compose.yaml` that provisions a working stack:

```bash
docker compose up -d
```

- Grafana: http://localhost:3000 (admin/admin)
- ArcadeDB: http://localhost:2480 (root/arcadedb)
- Data source is auto-provisioned and points at a `MovieRatings` database.

`MovieRatings` is not seeded by the compose stack. Create and populate it once via the ArcadeDB Studio at http://localhost:2480 (use the bundled MovieRatings sample importer), or run the equivalent setup script before the data source can return results.

Suggested queries to exercise each mode are listed in `README.md`.

## Release ordering

The release workflow's plugin validator step fails if `info.screenshots` entries in `src/plugin.json` reference files that do not exist in the build output. Screenshots must be captured and committed to `src/img/screenshots/` before any tag is pushed. See the implementation plan, Task 5, for the capture procedure.

## Validator warnings

Document any warning the validator emits that we deliberately accept. Add an entry per warning:

- _none currently_

## Release log

Track tag iterations during the review cycle here:

- `v1.0.0-beta.1` - initial submission

## Useful links

- Release workflow runs: https://github.com/ArcadeData/arcadedb-grafana-datasource/actions/workflows/release.yml
- GitHub releases: https://github.com/ArcadeData/arcadedb-grafana-datasource/releases
- Catalog listing (once approved): https://grafana.com/grafana/plugins/arcadedb-arcadedb-datasource
