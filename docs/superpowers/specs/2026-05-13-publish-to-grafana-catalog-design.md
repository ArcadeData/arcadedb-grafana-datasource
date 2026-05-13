# Publish ArcadeDB Data Source to the Grafana Plugin Catalog

Date: 2026-05-13
Author: r.franchini@arcadedata.com
Status: Approved for planning

## Goal

Submit `arcadedb-arcadedb-datasource` to the Grafana Plugin Catalog at version `1.0.0-beta.1`. After the first review cycle clears, tag `1.0.0` for GA.

## Done means

- A signed, validator-clean release zip is produced automatically when a `v*` tag is pushed to `main`.
- The plugin appears in the Grafana Plugin Catalog and can be installed via `grafana-cli plugins install arcadedb-arcadedb-datasource`.
- Reviewers at Grafana can install the plugin and verify it against an ArcadeDB instance using documented steps.

## Out of scope

- New plugin features.
- Self-hosted distribution channels (private signing, internal mirror, etc.).
- A separate marketing or docs site. The existing `docs.arcadedb.com` is the documentation target.
- Backporting to older Grafana versions than the `>=12.3.0` already declared.

## Constraints and decisions

- **Distribution route:** Grafana Plugin Catalog only.
- **First published version:** `1.0.0-beta.1`, promoted to `1.0.0` after first acceptance cycle.
- **Signing identity:** Arcade Data Ltd org account on grafana.com. Token stored in the GitHub repo secret `GRAFANA_ACCESS_POLICY_TOKEN` with the `plugins:write` scope.
- **Screenshots:** captured locally via a Playwright spec driving the existing `docker compose` stack; committed as PNGs.
- **Workflow shape:** one PR that lands all five work units, then a single `v1.0.0-beta.1` tag.

## Work breakdown

Five units inside a single PR. Listed in implementation order.

### Unit 1 — Repo hygiene

- Add `dist/`, `*.zip`, `*.sha1`, `MANIFEST.txt` to `.gitignore`.
- `git rm -r --cached dist/` to remove the committed build artifacts.
- Reconcile the Grafana version requirement. `src/plugin.json` declares `grafanaDependency: ">=12.3.0"`; `README.md` says "Grafana 10.0 or later". Source of truth is `plugin.json`. Update the README.
- Replace the placeholder `CHANGELOG.md` body with a real `1.0.0-beta.1` entry covering features, requirements, and install instructions.

Inputs: none.
Outputs: cleaner repo state. No behavior change.

### Unit 2 — Cross-platform backend build verification

- Confirm `Magefile.go` produces binaries for: linux amd64, linux arm64, darwin amd64, darwin arm64, windows amd64. These match the catalog's minimum platform matrix.
- If any platform is missing, add the corresponding mage target.

Inputs: existing `Magefile.go`.
Outputs: verified or extended build targets. Likely a no-op based on current `CLAUDE.md`.

### Unit 3 — Release workflow

New file: `.github/workflows/release.yml`.

Triggers:
- `push` to a tag matching `v*` — produces a real release.
- `workflow_dispatch` with input `dry_run: true` — produces a build artifact but skips the GitHub release step. Used for exercising the pipeline from the PR branch before any tag is cut.

Steps:

1. Checkout, `setup-node` (Node 25 to match CI), `setup-go` (from `go.mod`).
2. Install dependencies: `npm ci`.
3. Quality gates: `npm run typecheck`, `npm run lint`, `npm run test:ci`, `go test ./pkg/...`.
4. Frontend build: `npm run build` (emits to `dist/`).
5. Backend build: `mage -v build:linux build:darwin build:windows`.
6. Substitute placeholders in `dist/plugin.json`:
   - `%VERSION%` → tag name with leading `v` stripped (e.g. `1.0.0-beta.1`).
   - `%TODAY%` → UTC `YYYY-MM-DD`.
7. Sign: `npx @grafana/sign-plugin@latest`, with env `GRAFANA_ACCESS_POLICY_TOKEN`. Produces `dist/MANIFEST.txt`.
8. Package: rename `dist/` into a temp directory as `arcadedb-arcadedb-datasource/`, zip it as `arcadedb-arcadedb-datasource-<version>.zip`.
9. Validate: `npx @grafana/plugin-validator@latest -sourceCodeUri file://. <zip>`. Fails the build on any error. Warnings are surfaced in the step summary; specific known-safe warning IDs may be allowlisted via a workflow input if needed.
10. Checksum: `sha1sum <zip> > <zip>.sha1`.
11. Release: on tag-triggered runs only, use `softprops/action-gh-release@v2` to publish the zip, sha1, and a changelog excerpt for the tag.

Secrets required: `GRAFANA_ACCESS_POLICY_TOKEN`.

### Unit 4 — Screenshot capture

New file: `e2e/screenshots.spec.ts` (Playwright). Uses the existing `docker compose up` stack with the auto-provisioned data source.

Captures five PNGs into `src/img/screenshots/`:

1. `config.png` — data source configuration page populated with a working ArcadeDB URL.
2. `timeseries.png` — Time Series mode showing the visual builder and a result panel.
3. `sql.png` — SQL mode editor and result panel.
4. `cypher.png` — Cypher mode rendering a Node Graph panel.
5. `gremlin.png` — Gremlin mode rendering a Node Graph panel.

Wire `info.screenshots` in `src/plugin.json` with entries for each:

```json
"screenshots": [
  {"name": "Configuration", "path": "img/screenshots/config.png"},
  {"name": "Time Series", "path": "img/screenshots/timeseries.png"},
  {"name": "SQL", "path": "img/screenshots/sql.png"},
  {"name": "Cypher Node Graph", "path": "img/screenshots/cypher.png"},
  {"name": "Gremlin Node Graph", "path": "img/screenshots/gremlin.png"}
]
```

Add an `npm run screenshots` script that runs the Playwright spec. Run manually; not part of CI. Generated PNGs are committed to git.

### Unit 5 — Submission notes

New file: `docs/SUBMISSION.md`. Not bundled in the zip. Captures everything we need during the catalog review back-and-forth:

- grafana.com submission URL and the org account used.
- Contact email: r.franchini@arcadedata.com.
- Reviewer setup: `docker compose up -d`, default Grafana credentials, sample queries to try.
- Known validator warnings (if any) and why they are accepted.
- Links to the GitHub release and the release workflow run.

## Verification

Before tagging `v1.0.0-beta.1`:

- All CI checks green on the PR.
- One `workflow_dispatch` run from the PR branch with `dry_run: true`. Inspect the artifact:
  - Zip layout: top-level `arcadedb-arcadedb-datasource/` directory, contains `plugin.json`, `module.js`, `MANIFEST.txt`, backend binaries in `gpx_arcadedb_*`.
  - `plugin.json` has real version and date (placeholders substituted).
  - `MANIFEST.txt` is present and references the expected files.
  - Validator output has zero errors.
- Local install: download dry-run zip, extract into a local Grafana plugins directory, restart Grafana, configure the data source, run one query in each of the four modes.

After verification, push the tag. The same workflow runs in non-dry mode and creates the GitHub release.

## Risks

| Risk | Mitigation |
| --- | --- |
| Signing token not yet provisioned when tagging | Hold tag until the secret is in place; dry-run path skips signing for earlier exercises. |
| Validator flags an unfixable issue | Document in `docs/SUBMISSION.md` and ask the reviewer for guidance during review. |
| Screenshots drift as the UI evolves | `npm run screenshots` re-runs the capture against the current build. Treat as a maintenance command, not CI. |
| `>=12.3.0` is too narrow or too broad | The catalog reviewer typically tests against the declared minimum and current stable. They will request a change if needed. |
| Catalog review surfaces unexpected requirements | Iterate by tagging `v1.0.0-beta.2`, `v1.0.0-beta.3`, etc. Each iteration is one tag push. |

## Testing strategy

- Frontend and backend tests run on every PR via existing `.github/workflows/ci.yml`.
- Release workflow correctness is verified by `workflow_dispatch` dry runs before any tag is cut.
- Screenshot capture (`e2e/screenshots.spec.ts`) is manual; it depends on a live stack and is too coupled to UI specifics to gate CI on.
- Post-release smoke test: install the published zip into a local Grafana, run one query in each mode, confirm node graph renders.

## Publishing flow

1. Land the single PR with all five units to `main`.
2. Confirm `GRAFANA_ACCESS_POLICY_TOKEN` secret is set in the repo.
3. `git tag v1.0.0-beta.1 && git push origin v1.0.0-beta.1`.
4. Release workflow runs end-to-end, produces a GitHub release with the signed zip.
5. Submit the zip URL via grafana.com → My Plugins → Submit New Plugin.
6. Address reviewer feedback by tagging successive `v1.0.0-beta.N` releases.
7. After acceptance, tag `v1.0.0` for GA.
