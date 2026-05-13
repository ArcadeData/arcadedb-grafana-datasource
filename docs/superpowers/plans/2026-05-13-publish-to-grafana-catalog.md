# Publish to Grafana Plugin Catalog — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Land all repo changes needed to ship `arcadedb-arcadedb-datasource@1.0.0-beta.1` to the Grafana Plugin Catalog via a single PR, then tag the release.

**Architecture:** One PR containing five units of change (hygiene, magefile verify, release workflow, screenshots, submission notes), followed by a single signed tag that triggers the release workflow.

**Tech Stack:** GitHub Actions, Node 25, Go (from `go.mod`), Mage with `grafana-plugin-sdk-go/build`, `@grafana/sign-plugin`, `@grafana/plugin-validator`, Playwright.

**Spec:** `docs/superpowers/specs/2026-05-13-publish-to-grafana-catalog-design.md`

---

## File Structure

Files this plan creates or modifies:

| File | Action | Responsibility |
| --- | --- | --- |
| `.gitignore` | Modify | Add `*.zip`, `*.sha1`, `MANIFEST.txt` |
| `dist/` (tracked files) | Untrack | Remove committed build artifacts |
| `README.md` | Modify | Reconcile Grafana version requirement |
| `CHANGELOG.md` | Modify | Replace stub with real `1.0.0-beta.1` entry |
| `src/plugin.json` | Modify | Add `info.screenshots` array |
| `package.json` | Modify | Add `screenshots` script, `@playwright/test` devDependency |
| `playwright.config.ts` | Create | Playwright config pinned to local stack |
| `e2e/screenshots.spec.ts` | Create | Playwright spec that captures the five PNGs |
| `src/img/screenshots/*.png` | Create | 5 binary PNGs (config, ts, sql, cypher, gremlin) |
| `.github/workflows/release.yml` | Create | Tag/dispatch-driven build, sign, validate, publish |
| `docs/SUBMISSION.md` | Create | Catalog submission notes for reviewers |

---

## Task 1: Repo hygiene

**Files:**
- Modify: `.gitignore`
- Untrack: `dist/CHANGELOG.md`, `dist/LICENSE`, `dist/README.md`, `dist/module.js`, `dist/module.js.map`, `dist/plugin.json`, `dist/img/`
- Modify: `README.md:47`
- Modify: `CHANGELOG.md` (full replace)

- [ ] **Step 1: Add release artifact patterns to `.gitignore`**

Append to `.gitignore` (the file already excludes `dist/`, but does not exclude release packaging outputs):

```gitignore

# Release packaging
*.zip
*.sha1
MANIFEST.txt
```

- [ ] **Step 2: Untrack the committed `dist/` directory**

Run:

```bash
git rm -r --cached dist/
```

Expected: lines like `rm 'dist/plugin.json'` for each file under `dist/`.

- [ ] **Step 3: Verify dist is no longer tracked**

Run:

```bash
git ls-files dist/
```

Expected: no output.

- [ ] **Step 4: Fix the Grafana version in `README.md`**

Open `README.md`, line 47. Replace:

```
- Grafana 10.0 or later
```

with:

```
- Grafana 12.3.0 or later
```

- [ ] **Step 5: Rewrite `CHANGELOG.md` with a real entry**

Replace the entire contents of `CHANGELOG.md` with:

```markdown
# Changelog

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
```

- [ ] **Step 6: Sanity-check the working tree**

Run:

```bash
git status --short
```

Expected output includes:
- `M  .gitignore`
- `D  dist/CHANGELOG.md` (and the rest of dist/* as deleted)
- `M  README.md`
- `M  CHANGELOG.md`

- [ ] **Step 7: Commit**

```bash
git add .gitignore README.md CHANGELOG.md
git add -u dist/
git commit -m "chore: prepare repo for catalog submission"
```

---

## Task 2: Verify Magefile platform coverage

**Files:**
- Read-only: `Magefile.go`

The `Magefile.go` uses `build.BuildAll` from `github.com/grafana/grafana-plugin-sdk-go/build`. `BuildAll` builds for the platforms Grafana's catalog requires: linux amd64+arm64+arm, darwin amd64+arm64, windows amd64. This task verifies that locally; no edits expected.

- [ ] **Step 1: Build all platforms locally**

Run:

```bash
mage -v
```

Expected: completes without error. The default target is `BuildAll` (see `Magefile.go:9`).

- [ ] **Step 2: List produced binaries**

Run:

```bash
ls dist/gpx_arcadedb_*
```

Expected output contains at minimum:
- `dist/gpx_arcadedb_linux_amd64`
- `dist/gpx_arcadedb_linux_arm64`
- `dist/gpx_arcadedb_linux_arm`
- `dist/gpx_arcadedb_darwin_amd64`
- `dist/gpx_arcadedb_darwin_arm64`
- `dist/gpx_arcadedb_windows_amd64.exe`

- [ ] **Step 3: If any platform is missing**

If the listing in Step 2 lacks any platform above, extend `Magefile.go` with an explicit aggregate target. Otherwise skip to Step 4.

Replace the file contents with:

```go
//go:build mage
// +build mage

package main

import (
	// mage:import
	build "github.com/grafana/grafana-plugin-sdk-go/build"
)

// Default builds all required platforms for the Grafana catalog.
var Default = build.BuildAll
```

(This is the same content the file already has; only edit if `BuildAll` in the SDK has been narrowed in a future version. If the listing was complete, no commit is needed.)

- [ ] **Step 4: Clean up local build output**

`dist/` is gitignored, but remove it to keep the workspace clean:

```bash
rm -rf dist/
```

Expected: no output. No commit produced by this task unless `Magefile.go` had to change.

---

## Task 3: Install Playwright and create config

**Files:**
- Modify: `package.json`
- Create: `playwright.config.ts`

- [ ] **Step 1: Install `@playwright/test` as a dev dependency**

Run:

```bash
npm install -D @playwright/test@latest
```

Expected: `package.json` and `package-lock.json` updated; `@playwright/test` appears under `devDependencies`.

- [ ] **Step 2: Install browser binaries**

Run:

```bash
npx playwright install chromium
```

Expected: chromium downloaded to `~/Library/Caches/ms-playwright/` (macOS) or equivalent.

- [ ] **Step 3: Create `playwright.config.ts`**

Create the file with:

```typescript
import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  timeout: 60_000,
  fullyParallel: false,
  workers: 1,
  reporter: 'list',
  use: {
    baseURL: 'http://localhost:3000',
    viewport: { width: 1440, height: 900 },
    httpCredentials: { username: 'admin', password: 'admin' },
    screenshot: 'off',
  },
  projects: [
    {
      name: 'chromium',
      use: { browserName: 'chromium' },
    },
  ],
});
```

- [ ] **Step 4: Commit**

```bash
git add package.json package-lock.json playwright.config.ts
git commit -m "chore: add playwright for screenshot capture"
```

---

## Task 4: Write the screenshot capture spec

**Files:**
- Create: `e2e/screenshots.spec.ts`
- Modify: `package.json` (add `screenshots` script)

The spec assumes the local `docker compose` stack is running and the `MovieRatings` database has been imported in ArcadeDB. The spec does not seed data; it only drives the UI.

- [ ] **Step 1: Create `e2e/screenshots.spec.ts`**

Create the file with:

```typescript
import { test, Page } from '@playwright/test';
import { mkdirSync } from 'fs';
import { resolve } from 'path';

const OUT_DIR = resolve(__dirname, '../src/img/screenshots');

test.beforeAll(() => {
  mkdirSync(OUT_DIR, { recursive: true });
});

async function login(page: Page) {
  await page.goto('/login');
  await page.fill('input[name="user"]', 'admin');
  await page.fill('input[name="password"]', 'admin');
  await page.click('button[type="submit"]');
  // Skip the "change password" prompt if it appears.
  const skip = page.getByRole('link', { name: /skip/i });
  if (await skip.isVisible().catch(() => false)) {
    await skip.click();
  }
  await page.waitForURL(/\/$|\/home/);
}

async function openExplore(page: Page) {
  await page.goto('/explore');
  // Select the ArcadeDB data source.
  const dsPicker = page.getByTestId('data-testid Data source picker select container');
  await dsPicker.click();
  await page.getByText('ArcadeDB', { exact: false }).first().click();
}

async function runQuery(page: Page) {
  await page.getByRole('button', { name: /run query/i }).click();
  await page.waitForTimeout(2000);
}

test('config screenshot', async ({ page }) => {
  await login(page);
  await page.goto('/connections/datasources');
  await page.getByText('ArcadeDB', { exact: false }).first().click();
  await page.waitForLoadState('networkidle');
  await page.screenshot({ path: `${OUT_DIR}/config.png`, fullPage: false });
});

test('time series screenshot', async ({ page }) => {
  await login(page);
  await openExplore(page);
  // Time Series tab is default for this plugin.
  // Pick first available type and field via the visual builder.
  // Tweak selectors to match QueryEditor.tsx if needed.
  await page.waitForTimeout(1500);
  await runQuery(page);
  await page.screenshot({ path: `${OUT_DIR}/timeseries.png`, fullPage: false });
});

test('sql screenshot', async ({ page }) => {
  await login(page);
  await openExplore(page);
  await page.getByRole('tab', { name: /sql/i }).click();
  // Monaco editor; type a query that hits MovieRatings sample data.
  await page.locator('.monaco-editor').first().click();
  await page.keyboard.type('SELECT FROM Movie LIMIT 10');
  await runQuery(page);
  await page.screenshot({ path: `${OUT_DIR}/sql.png`, fullPage: false });
});

test('cypher screenshot', async ({ page }) => {
  await login(page);
  await openExplore(page);
  await page.getByRole('tab', { name: /cypher/i }).click();
  await page.locator('.monaco-editor').first().click();
  await page.keyboard.type('MATCH (m:Movie)-[r]-(n) RETURN m, r, n LIMIT 25');
  await runQuery(page);
  await page.screenshot({ path: `${OUT_DIR}/cypher.png`, fullPage: false });
});

test('gremlin screenshot', async ({ page }) => {
  await login(page);
  await openExplore(page);
  await page.getByRole('tab', { name: /gremlin/i }).click();
  await page.locator('.monaco-editor').first().click();
  await page.keyboard.type('g.V().limit(25)');
  await runQuery(page);
  await page.screenshot({ path: `${OUT_DIR}/gremlin.png`, fullPage: false });
});
```

The selectors are best-effort; if a test fails because a UI element is named differently, update the selector to match what the running app actually shows. The engineer executing Task 5 will fix any drift.

- [ ] **Step 2: Add the `screenshots` npm script**

Edit `package.json`. In the `scripts` block, add the `screenshots` entry between `e2e` and `sign`:

```json
    "e2e": "playwright test",
    "screenshots": "playwright test e2e/screenshots.spec.ts",
    "sign": "npx --yes @grafana/sign-plugin@latest"
```

- [ ] **Step 3: Commit**

```bash
git add e2e/screenshots.spec.ts package.json
git commit -m "feat: add playwright spec for catalog screenshots"
```

---

## Task 5: Capture screenshots against the local stack

**Files:**
- Create (binary): `src/img/screenshots/config.png`
- Create (binary): `src/img/screenshots/timeseries.png`
- Create (binary): `src/img/screenshots/sql.png`
- Create (binary): `src/img/screenshots/cypher.png`
- Create (binary): `src/img/screenshots/gremlin.png`

This is a manual capture step. The engineer needs Docker running and may need to fix selectors. PNGs are committed; CI does not regenerate them.

- [ ] **Step 1: Build the plugin so the dev stack picks it up**

```bash
npm run build && mage -v
```

Expected: completes without error; `dist/` is populated.

- [ ] **Step 2: Start the dev stack**

```bash
docker compose up -d
```

Expected: `grafana` and `arcadedb` containers are healthy.

- [ ] **Step 3: Confirm Grafana is reachable**

```bash
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:3000/login
```

Expected: `200`.

- [ ] **Step 4: Ensure ArcadeDB has sample data**

The provisioned data source points at the `MovieRatings` database. Check it exists:

```bash
curl -s -u root:arcadedb http://localhost:2480/api/v1/databases
```

If `MovieRatings` is not in the list, import it (ArcadeDB ships an importer):

```bash
curl -s -u root:arcadedb -X POST http://localhost:2480/api/v1/server \
  -H 'Content-Type: application/json' \
  -d '{"command":"create database MovieRatings"}'
# Then import sample data via your preferred method, e.g. studio at http://localhost:2480
```

- [ ] **Step 5: Run the screenshot spec**

```bash
npm run screenshots
```

Expected: five PNGs appear in `src/img/screenshots/`. If any test fails because of a selector mismatch, edit `e2e/screenshots.spec.ts` to match what the UI actually shows, then re-run.

- [ ] **Step 6: Eyeball the PNGs**

Open each of the five files and confirm they show useful content (data visible, no error overlay, no half-loaded panels). Re-run if needed.

- [ ] **Step 7: Stop the dev stack**

```bash
docker compose down
```

- [ ] **Step 8: Commit the screenshots**

```bash
git add src/img/screenshots/
git commit -m "feat: capture catalog screenshots"
```

---

## Task 6: Wire screenshots into `plugin.json`

**Files:**
- Modify: `src/plugin.json`

- [ ] **Step 1: Add `screenshots` to the `info` block**

In `src/plugin.json`, inside the `info` object (around the `logos` field), insert:

```json
    "screenshots": [
      { "name": "Configuration",       "path": "img/screenshots/config.png" },
      { "name": "Time Series",         "path": "img/screenshots/timeseries.png" },
      { "name": "SQL",                 "path": "img/screenshots/sql.png" },
      { "name": "Cypher Node Graph",   "path": "img/screenshots/cypher.png" },
      { "name": "Gremlin Node Graph",  "path": "img/screenshots/gremlin.png" }
    ],
```

Place it between `logos` and `links`. Watch comma placement.

- [ ] **Step 2: Validate JSON parses**

```bash
node -e "JSON.parse(require('fs').readFileSync('src/plugin.json','utf8'))" && echo OK
```

Expected: `OK`.

- [ ] **Step 3: Commit**

```bash
git add src/plugin.json
git commit -m "feat: reference screenshots in plugin manifest"
```

---

## Task 7: Release workflow

**Files:**
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Create the workflow**

Create `.github/workflows/release.yml` with:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'
  workflow_dispatch:
    inputs:
      dry_run:
        description: 'Skip GitHub release upload (build/sign/validate only)'
        type: boolean
        default: true

permissions:
  contents: write

jobs:
  release:
    name: Build, sign, and release
    runs-on: ubuntu-latest
    env:
      GRAFANA_ACCESS_POLICY_TOKEN: ${{ secrets.GRAFANA_ACCESS_POLICY_TOKEN }}
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Node
        uses: actions/setup-node@v4
        with:
          node-version: '25'
          cache: 'npm'

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Resolve version
        id: version
        run: |
          if [[ "${GITHUB_REF}" == refs/tags/* ]]; then
            VERSION="${GITHUB_REF#refs/tags/v}"
          else
            VERSION="0.0.0-dev.${GITHUB_SHA::7}"
          fi
          echo "version=${VERSION}" >> "${GITHUB_OUTPUT}"
          echo "today=$(date -u +%Y-%m-%d)" >> "${GITHUB_OUTPUT}"

      - name: Install npm deps
        run: npm ci

      - name: Frontend gates
        run: |
          npm run typecheck
          npm run lint
          npm run test:ci

      - name: Backend tests
        run: go test -race -v ./pkg/... -timeout 300s

      - name: Frontend build
        run: npm run build

      - name: Backend build
        run: |
          go install github.com/magefile/mage@latest
          mage -v

      - name: Substitute version and date in plugin.json
        run: |
          jq --arg v "${{ steps.version.outputs.version }}" \
             --arg d "${{ steps.version.outputs.today }}" \
             '.info.version = $v | .info.updated = $d' \
             dist/plugin.json > dist/plugin.json.tmp
          mv dist/plugin.json.tmp dist/plugin.json

      - name: Sign plugin
        if: env.GRAFANA_ACCESS_POLICY_TOKEN != ''
        run: npx --yes @grafana/sign-plugin@latest

      - name: Package zip
        id: package
        run: |
          PLUGIN_ID=$(jq -r .id dist/plugin.json)
          VERSION="${{ steps.version.outputs.version }}"
          STAGING="$(mktemp -d)/${PLUGIN_ID}"
          mkdir -p "${STAGING}"
          cp -r dist/. "${STAGING}/"
          ZIP_NAME="${PLUGIN_ID}-${VERSION}.zip"
          (cd "$(dirname "${STAGING}")" && zip -r "${GITHUB_WORKSPACE}/${ZIP_NAME}" "${PLUGIN_ID}")
          echo "zip=${ZIP_NAME}" >> "${GITHUB_OUTPUT}"

      - name: Validate plugin
        run: |
          npx --yes @grafana/plugin-validator@latest \
            -sourceCodeUri file://. \
            "${{ steps.package.outputs.zip }}"

      - name: Generate sha1
        run: sha1sum "${{ steps.package.outputs.zip }}" > "${{ steps.package.outputs.zip }}.sha1"

      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: plugin-${{ steps.version.outputs.version }}
          path: |
            ${{ steps.package.outputs.zip }}
            ${{ steps.package.outputs.zip }}.sha1
          if-no-files-found: error

      - name: Create GitHub release
        if: startsWith(github.ref, 'refs/tags/v') && (github.event_name != 'workflow_dispatch' || inputs.dry_run == false)
        uses: softprops/action-gh-release@v2
        with:
          files: |
            ${{ steps.package.outputs.zip }}
            ${{ steps.package.outputs.zip }}.sha1
          body_path: CHANGELOG.md
          draft: false
          prerelease: ${{ contains(steps.version.outputs.version, '-') }}
```

- [ ] **Step 2: Lint the workflow YAML**

```bash
python3 -c "import yaml,sys; yaml.safe_load(open('.github/workflows/release.yml'))" && echo OK
```

Expected: `OK`.

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "ci: add release workflow with sign+validate"
```

---

## Task 8: Submission notes

**Files:**
- Create: `docs/SUBMISSION.md`

- [ ] **Step 1: Create the document**

Create `docs/SUBMISSION.md` with:

```markdown
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
- Data source is auto-provisioned against the `MovieRatings` sample database.

Suggested queries to exercise each mode are listed in `README.md`.

## Validator warnings

Document any warning the validator emits that we deliberately accept. Add an entry per warning:

- _none currently_

## Release log

Track tag iterations during the review cycle here:

- `v1.0.0-beta.1` — initial submission

## Useful links

- Release workflow runs: https://github.com/ArcadeData/arcadedb-grafana-datasource/actions/workflows/release.yml
- GitHub releases: https://github.com/ArcadeData/arcadedb-grafana-datasource/releases
- Catalog listing (once approved): https://grafana.com/grafana/plugins/arcadedb-arcadedb-datasource
```

- [ ] **Step 2: Commit**

```bash
git add docs/SUBMISSION.md
git commit -m "docs: add catalog submission notes"
```

---

## Task 9: Dry-run the release workflow from the PR branch

**Files:** none (verification only).

- [ ] **Step 1: Push the branch**

```bash
git push -u origin HEAD
```

- [ ] **Step 2: Trigger `workflow_dispatch` with `dry_run: true`**

Either via GitHub UI (Actions -> Release -> Run workflow) or:

```bash
gh workflow run release.yml -f dry_run=true
```

Expected: a new workflow run starts.

- [ ] **Step 3: Wait for completion and inspect the artifact**

```bash
gh run watch
gh run download
```

Expected: a directory per artifact (e.g. `plugin-0.0.0-dev.abc1234/`) containing a zip and a `.sha1` file.

- [ ] **Step 4: Verify zip layout**

```bash
unzip -l arcadedb-arcadedb-datasource-*.zip | head -40
```

Expected:
- Top-level directory `arcadedb-arcadedb-datasource/`
- Contains `plugin.json`, `module.js`, `MANIFEST.txt`, `img/`, backend binaries `gpx_arcadedb_*`

- [ ] **Step 5: Verify `plugin.json` substitutions**

```bash
unzip -p arcadedb-arcadedb-datasource-*.zip arcadedb-arcadedb-datasource/plugin.json | jq '.info.version, .info.updated'
```

Expected: a real version string (not `%VERSION%`) and a date (not `%TODAY%`).

- [ ] **Step 6: Confirm validator output is clean**

In the workflow run summary, the "Validate plugin" step should report zero errors. Warnings, if any, get logged in `docs/SUBMISSION.md` (Task 8).

- [ ] **Step 7: If anything failed**

Fix the issue in a new commit on the same branch, push, and re-run the dispatch. Do not proceed until the dry run is clean.

---

## Task 10: Merge and tag

**Files:** none.

- [ ] **Step 1: Open a PR**

```bash
gh pr create --title "Prepare for Grafana catalog submission" \
  --body "$(cat <<'EOF'
## Summary
- Repo hygiene (gitignore release artifacts, untrack dist/, fix Grafana version in README, fill CHANGELOG)
- Playwright-based screenshot capture for catalog listing
- plugin.json wired with screenshots
- New release workflow that builds, signs, validates, packages, and publishes to GitHub releases
- Submission notes for the catalog review cycle

## Test plan
- [x] Frontend CI (typecheck, lint, jest)
- [x] Backend CI (go test -race)
- [x] Release workflow dry-run produces a clean signed zip and a passing validator report
- [ ] Reviewers: smoke-test all four query modes against the dev stack
EOF
)"
```

- [ ] **Step 2: After approval, merge to `main`**

Use the project's normal merge strategy.

- [ ] **Step 3: Confirm the signing secret is set**

```bash
gh secret list | grep GRAFANA_ACCESS_POLICY_TOKEN
```

Expected: the secret is present. If not, set it before tagging:

```bash
gh secret set GRAFANA_ACCESS_POLICY_TOKEN
```

(Paste the Arcade Data org access policy token from grafana.com when prompted.)

- [ ] **Step 4: Tag the release**

```bash
git checkout main
git pull
git tag v1.0.0-beta.1
git push origin v1.0.0-beta.1
```

- [ ] **Step 5: Watch the release workflow**

```bash
gh run watch
```

Expected: workflow completes, a GitHub release `v1.0.0-beta.1` is created with the signed zip and sha1.

- [ ] **Step 6: Submit to the catalog**

Go to https://grafana.com -> My Plugins -> Submit New Plugin. Provide the GitHub release zip URL. Update `docs/SUBMISSION.md` with the submission outcome.

---

## Self-review notes

- **Spec coverage:** Unit 1 → Task 1; Unit 2 → Task 2; Unit 3 → Task 7; Unit 4 → Tasks 3, 4, 5, 6; Unit 5 → Task 8. Verification step from the spec → Task 9. Publishing flow from the spec → Task 10.
- **Placeholder scan:** All steps have concrete commands or file content. The only "engineer judgment" point is selector drift in `e2e/screenshots.spec.ts` (Task 4 step 1), which is explicitly called out as expected to need fixing during Task 5.
- **Type consistency:** No code types crossing tasks. `plugin.json` `info.screenshots` shape (Task 6) matches the paths produced by `e2e/screenshots.spec.ts` (Task 4).
- **Known risks acknowledged in plan:** signing token absence handled by `if: env.GRAFANA_ACCESS_POLICY_TOKEN != ''` in the workflow; dry-run path skips release upload; sample data import is a manual prerequisite documented in Task 5 Step 4.
