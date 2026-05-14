# Grafana Plugin Catalog Submission

Internal notes for shipping `arcadedb-arcadedb-datasource` to the catalog. Not bundled in the release zip.

## Identifiers

- Plugin ID: `arcadedb-arcadedb-datasource`
- Org account on grafana.com: Arcade Data Ltd
- Submission URL: https://grafana.com/auth/sign-in -> My Plugins -> Submit New Plugin
- Signing token secret: `GRAFANA_ACCESS_POLICY_TOKEN` (scope `plugins:write`)
- Primary contact: r.franchini@arcadedata.com

## Submission flow (first release)

The catalog accepts a built zip URL plus a SHA1 plus a source-code tree URL. To produce these:

1. Make sure `main` has the validator fixes (PR #46) and signing soft-fail (PR #47) merged.
2. Tag and push:

   ```bash
   git checkout main && git pull
   git tag v1.0.0-beta.1
   git push origin v1.0.0-beta.1
   ```

3. The `release.yml` workflow runs. Signing 403s on the first submission because the `arcadedb-` ID prefix is not yet registered to the grafana.com org; the step is `continue-on-error: true` so the workflow still publishes the unsigned zip.
4. Confirm the GitHub release at `https://github.com/ArcadeData/arcadedb-grafana-datasource/releases/tag/v1.0.0-beta.1` contains:
   - `arcadedb-arcadedb-datasource-1.0.0-beta.1.zip`
   - `arcadedb-arcadedb-datasource-1.0.0-beta.1.zip.sha1`

## Catalog submission form (Create Plugin Submission)

Field-by-field values for the form at grafana.com:

| Field | Value |
| --- | --- |
| How do you want to list your plugin? | Public (free) |
| Plugin URL (zip file) | `https://github.com/ArcadeData/arcadedb-grafana-datasource/releases/download/v1.0.0-beta.1/arcadedb-arcadedb-datasource-1.0.0-beta.1.zip` |
| MD5 or SHA1 | `https://github.com/ArcadeData/arcadedb-grafana-datasource/releases/download/v1.0.0-beta.1/arcadedb-arcadedb-datasource-1.0.0-beta.1.zip.sha1` |
| Source code URL | `https://github.com/ArcadeData/arcadedb-grafana-datasource/tree/v1.0.0-beta.1` |
| Provisioning provided for test environment | Yes |
| Are you affiliated with the project/product the plugin integrates with? | Yes (Arcade Data Ltd publishes both ArcadeDB and this plugin) |
| Does the plugin integrate with a commercial product? | No (ArcadeDB is open source) |

### Testing guidance (paste into the form)

```
The repo includes a docker compose stack that provisions Grafana and ArcadeDB
with the data source pre-configured:

  git clone https://github.com/ArcadeData/arcadedb-grafana-datasource
  cd arcadedb-grafana-datasource
  docker compose up -d

Grafana: http://localhost:3000 (admin/admin)
ArcadeDB: http://localhost:2480 (root/arcadedb)
Data source: provisioned via provisioning/datasources/arcadedb.yaml against
the MovieRatings sample database.

To populate MovieRatings:
  curl -u root:arcadedb -X POST http://localhost:2480/api/v1/server \
    -H 'Content-Type: application/json' \
    -d '{"command":"create database MovieRatings"}'

  curl -u root:arcadedb -X POST http://localhost:2480/api/v1/command/MovieRatings \
    -H 'Content-Type: application/json' \
    -d '{"language":"sql","command":"IMPORT DATABASE https://github.com/ArcadeData/arcadedb-datasets/raw/main/orientdb/MovieRatings.gz"}'

Then in Grafana, navigate to Explore, pick the ArcadeDB data source, and
exercise the four query modes:

  SQL:     SELECT FROM Movies LIMIT 10
  Cypher:  MATCH (m:Movies)<-[r:rated]-(u:Users) RETURN m, r, u LIMIT 25
           (enable the Node Graph toggle)
  Gremlin: g.V().limit(25)

Time Series mode requires registered time-series types in the target
database; MovieRatings has none, so the visual builder dropdowns are
empty against this sample. Time series queries work against any
ArcadeDB database that has registered time-series types via the
/api/v1/ts/{db}/grafana/metadata endpoint.

Submitted unsigned per the validator's "unsigned plugin" guidance.
Please entitle the arcadedb- ID prefix to the submitting org so
subsequent releases (v1.0.0 GA and onward) can be signed.
```

## Reviewer setup (compose-only summary)

The plugin ships with a `docker-compose.yaml` that provisions a working stack:

```bash
docker compose up -d
```

- Grafana: http://localhost:3000 (admin/admin)
- ArcadeDB: http://localhost:2480 (root/arcadedb)
- Data source is auto-provisioned and points at a `MovieRatings` database.

`MovieRatings` is not seeded by the compose stack. Create and populate it once via the ArcadeDB Studio at http://localhost:2480 (use the bundled MovieRatings sample importer), or via the curl commands in the testing guidance above.

## Release ordering

The release workflow's plugin validator step fails if `info.screenshots` entries in `src/plugin.json` reference files that do not exist in the build output. Screenshots must be captured and committed to `src/img/screenshots/` before any tag is pushed. See the implementation plan, Task 5, for the capture procedure.

## Validator warnings

Document any warning the validator emits that we deliberately accept. Add an entry per warning:

- `warning: unsigned plugin` - expected during first review. Grafana will sign subsequent releases after entitling the `arcadedb-` ID prefix to our org.

## Post-acceptance followups

Once Grafana accepts the plugin and entitles the org for signing:

- Remove `continue-on-error: true` from the Sign step in `.github/workflows/release.yml` so future signing failures fail the build.
- Tag `v1.0.0` for GA.

## Release log

Track tag iterations during the review cycle here:

- `v1.0.0-beta.1` - initial submission

## Useful links

- Release workflow runs: https://github.com/ArcadeData/arcadedb-grafana-datasource/actions/workflows/release.yml
- GitHub releases: https://github.com/ArcadeData/arcadedb-grafana-datasource/releases
- Catalog listing (once approved): https://grafana.com/grafana/plugins/arcadedb-arcadedb-datasource
