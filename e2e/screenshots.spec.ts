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
