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
  // After login Grafana may redirect to a password-change page. Wait for either
  // that page or the home page, then dismiss the password prompt if present.
  await page.waitForLoadState('networkidle');
  const skip = page.getByRole('button', { name: /^skip$/i });
  if (await skip.isVisible({ timeout: 5000 }).catch(() => false)) {
    await skip.click();
    await page.waitForLoadState('networkidle');
  }
}

async function openExplore(page: Page) {
  await page.goto('/explore');
  await page.waitForLoadState('networkidle');
  // Select the ArcadeDB data source if a picker is shown.
  const dsPicker = page.getByTestId('data-testid Data source picker select container');
  if (await dsPicker.isVisible({ timeout: 5000 }).catch(() => false)) {
    await dsPicker.click();
    await page.getByText('ArcadeDB', { exact: false }).first().click();
  }
  // Wait for the query editor to render.
  await page.waitForTimeout(1500);
}

async function selectMode(page: Page, label: string) {
  // The plugin's QueryEditor renders an InlineField with label "Mode" wrapping
  // a Grafana Select dropdown. Open it and click the option.
  const modeSelect = page.locator('[role="combobox"]').first();
  await modeSelect.click();
  await page.getByRole('option', { name: label, exact: true }).click();
  await page.waitForTimeout(500);
}

async function runQuery(page: Page) {
  await page.getByRole('button', { name: /run query/i }).click();
  await page.waitForTimeout(3000);
}

test('config screenshot', async ({ page }) => {
  await login(page);
  await page.goto('/connections/datasources');
  await page.getByText('ArcadeDB', { exact: false }).first().click();
  await page.waitForLoadState('networkidle');
  await page.screenshot({ path: `${OUT_DIR}/config.png`, fullPage: false });
});

test.skip('time series screenshot', async ({ page }) => {
  // Requires ArcadeDB time-series types to be registered in the target DB.
  // Re-enable when running against a database that has TS types so the
  // visual builder dropdowns render populated.
  await login(page);
  await openExplore(page);
  await selectMode(page, 'Time Series');
  await page.waitForTimeout(1500);
  await page.screenshot({ path: `${OUT_DIR}/timeseries.png`, fullPage: false });
});

test('sql screenshot', async ({ page }) => {
  await login(page);
  await openExplore(page);
  await selectMode(page, 'SQL');
  await page.locator('.monaco-editor').first().click();
  await page.keyboard.type('SELECT FROM Movies LIMIT 10');
  await runQuery(page);
  await page.screenshot({ path: `${OUT_DIR}/sql.png`, fullPage: false });
});

async function enableNodeGraph(page: Page) {
  // The InlineSwitch labelled "Node Graph" wraps an input[type=checkbox].
  // Set the checked state directly to avoid layered-element click issues.
  const checkbox = page.locator('input[type="checkbox"]').filter({
    has: page.locator('xpath=ancestor::*[contains(@class, "Switch") or contains(@class, "switch")]'),
  });
  // Fall back: find any checkbox following the Node Graph label.
  const nodeGraphCheckbox = page.locator('xpath=//label[normalize-space(.)="Node Graph"]/following::input[@type="checkbox"][1]');
  const target = (await nodeGraphCheckbox.count()) > 0 ? nodeGraphCheckbox : checkbox;
  await target.first().check({ force: true });
  await page.waitForTimeout(500);
}

test('cypher screenshot', async ({ page }) => {
  await login(page);
  await openExplore(page);
  await selectMode(page, 'Cypher');
  await enableNodeGraph(page);
  await page.locator('.monaco-editor').first().click();
  await page.keyboard.type('MATCH (m:Movies)<-[r:rated]-(u:Users) RETURN m, r, u LIMIT 25');
  await runQuery(page);
  await page.screenshot({ path: `${OUT_DIR}/cypher.png`, fullPage: false });
});

test('gremlin screenshot', async ({ page }) => {
  await login(page);
  await openExplore(page);
  await selectMode(page, 'Gremlin');
  await page.locator('.monaco-editor').first().click();
  await page.keyboard.type('g.V().limit(25)');
  await runQuery(page);
  await page.screenshot({ path: `${OUT_DIR}/gremlin.png`, fullPage: false });
});
