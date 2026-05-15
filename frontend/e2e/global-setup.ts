import { chromium, request } from '@playwright/test'
import path from 'path'
import fs from 'fs'

const API_BASE = process.env.PLAYWRIGHT_API_URL ?? 'http://localhost:8080'
const BASE_URL  = process.env.PLAYWRIGHT_BASE_URL ?? 'http://localhost:3000'

// ── Configurable credentials (overridden by env vars in CI) ──────────────────

const ADMIN_USERNAME = process.env.TEST_ADMIN_USERNAME ?? 'budi_santoso'
const ADMIN_PASSWORD = process.env.TEST_ADMIN_PASSWORD ?? 'Admin@Dana2025!'

const STAFF_USERNAME = process.env.TEST_STAFF_USERNAME ?? 'dewi_rahayu'
const STAFF_PASSWORD = process.env.TEST_STAFF_PASSWORD ?? 'Staff@Dana2025!'

// ── Saved auth-state paths (consumed by spec files via test.use) ─────────────

export const ADMIN_AUTH_STATE = path.join(__dirname, '.auth', 'admin.json')
export const STAFF_AUTH_STATE = path.join(__dirname, '.auth', 'staff.json')

// ── Test users to seed before every E2E run ──────────────────────────────────

const TEST_USERS = [
  {
    username: ADMIN_USERNAME,
    email:    `${ADMIN_USERNAME}@danasejahtera.id`,
    password: ADMIN_PASSWORD,
    role:     'admin',
  },
  {
    username: STAFF_USERNAME,
    email:    `${STAFF_USERNAME}@danasejahtera.id`,
    password: STAFF_PASSWORD,
    role:     'staff',
  },
]

// ── Helpers ───────────────────────────────────────────────────────────────────

function ensureDir(filePath: string) {
  const dir = path.dirname(filePath)
  if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true })
}

async function loginAndSaveState(
  browserType: typeof chromium,
  username: string,
  password: string,
  statePath: string,
): Promise<void> {
  const browser = await browserType.launch()
  const page    = await browser.newPage()

  await page.goto(`${BASE_URL}/login`)
  await page.waitForLoadState('networkidle')

  await page.fill('[placeholder="Masukkan username"]', username)
  await page.fill('[placeholder="Masukkan password"]', password)
  await page.click('button[type=submit]')
  await page.waitForURL('**/dashboard', { timeout: 15000 })

  await page.context().storageState({ path: statePath })
  await browser.close()
}

// ── Main setup ────────────────────────────────────────────────────────────────

export default async function globalSetup() {
  const apiCtx = await request.newContext({ baseURL: API_BASE })

  // 1. Reset to secure mode so all tests start from a known state.
  await apiCtx.put('/api/system/mode', {
    data:    { mode: 'secure' },
    headers: { 'Content-Type': 'application/json' },
  })

  // 2. Seed every test user (register endpoint is idempotent — 409 is benign).
  for (const u of TEST_USERS) {
    await apiCtx.post('/api/v1/auth/register', {
      data:    u,
      headers: { 'Content-Type': 'application/json' },
    })
    // Ignore response — 201 Created or 409 Conflict are both acceptable here.
  }

  await apiCtx.dispose()

  // 3. Login once per role and persist auth state for reuse.
  //    This keeps total login calls within the backend rate limit (5 req / 60 s).
  ensureDir(ADMIN_AUTH_STATE)
  ensureDir(STAFF_AUTH_STATE)

  await loginAndSaveState(chromium, ADMIN_USERNAME, ADMIN_PASSWORD, ADMIN_AUTH_STATE)
  await loginAndSaveState(chromium, STAFF_USERNAME, STAFF_PASSWORD, STAFF_AUTH_STATE)
}
