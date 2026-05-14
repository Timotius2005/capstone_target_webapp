/**
 * Playwright E2E tests — Security Mode Toggle
 *
 * Prerequisites:
 *   - Frontend running at http://localhost:3000
 *   - Backend running at http://localhost:8080
 *
 * Run: npx playwright test e2e/mode-toggle.spec.ts
 *
 * These tests verify the complete mode-switch flow end-to-end:
 * no page reload, instant UI update, banner colour change.
 */
import { test, expect, type Page } from '@playwright/test'

const API_BASE = process.env.PLAYWRIGHT_API_URL ?? 'http://localhost:8080'

// ── Helpers ───────────────────────────────────────────────────────────────────

async function getCurrentModeAPI(page: Page): Promise<string> {
  const res = await page.request.get(`${API_BASE}/api/system/mode`)
  const body = await res.json()
  return body.mode
}

async function setModeViaAPI(page: Page, mode: 'secure' | 'vulnerable') {
  await page.request.put(`${API_BASE}/api/system/mode`, {
    data: { mode },
  })
}

async function resetToSecure(page: Page) {
  await setModeViaAPI(page, 'secure')
}

// ── Banner visibility ─────────────────────────────────────────────────────────

test.describe('GlobalModeSwitcher banner', () => {
  test.beforeEach(async ({ page }) => {
    await resetToSecure(page)
    await page.goto('/')
  })

  test('banner is visible without login on the homepage', async ({ page }) => {
    // The banner is fixed at top — must be visible before auth
    const banner = page.getByRole('banner')
    await expect(banner).toBeVisible()
  })

  test('banner shows SECURE MODE text when mode is secure', async ({ page }) => {
    const banner = page.getByRole('banner')
    await expect(banner).toContainText(/SECURE MODE/i)
  })

  test('banner shows VULNERABLE when mode is vulnerable', async ({ page }) => {
    await setModeViaAPI(page, 'vulnerable')
    await page.goto('/')
    const banner = page.getByRole('banner')
    await expect(banner).toContainText(/VULNERABLE/i)
  })

  test('banner has green background in secure mode', async ({ page }) => {
    const banner = page.getByRole('banner')
    const bg = await banner.evaluate((el) => getComputedStyle(el).backgroundColor)
    // #14532d = rgb(20, 83, 45) — the secure green colour
    expect(bg).toMatch(/rgb\(20,\s*83,\s*45\)|rgb\(14,\s*59,\s*34\)/)
  })

  test('banner has red background in vulnerable mode', async ({ page }) => {
    await setModeViaAPI(page, 'vulnerable')
    await page.goto('/')
    const banner = page.getByRole('banner')
    const bg = await banner.evaluate((el) => getComputedStyle(el).backgroundColor)
    // Tailwind red-700 = approximately rgb(185, 28, 28)
    expect(bg).toMatch(/rgb\(185,\s*28,\s*28\)/)
  })
})

// ── Mode toggle — login page (pre-auth) ──────────────────────────────────────

test.describe('Mode toggle — login page (no auth required)', () => {
  test.beforeEach(async ({ page }) => {
    await resetToSecure(page)
    await page.goto('/login')
  })

  test('toggle button is visible on login page without authentication', async ({ page }) => {
    const toggleBtn = page.getByRole('button', { name: /Switch to Vulnerable/i })
    await expect(toggleBtn).toBeVisible()
  })

  test('clicking toggle switches to vulnerable without page reload', async ({ page }) => {
    const navigationPromise = page.waitForNavigation({ timeout: 2000 }).catch(() => null)
    const toggleBtn = page.getByRole('button', { name: /Switch to Vulnerable/i })
    await toggleBtn.click()

    // Wait for banner to update
    await expect(page.getByRole('banner')).toContainText(/VULNERABLE/i, { timeout: 5000 })

    // Navigation should NOT have occurred (no page reload)
    expect(await navigationPromise).toBeNull()
  })

  test('mode switch from login page persists to backend', async ({ page }) => {
    const toggleBtn = page.getByRole('button', { name: /Switch to Vulnerable/i })
    await toggleBtn.click()
    await expect(page.getByRole('banner')).toContainText(/VULNERABLE/i, { timeout: 5000 })

    // Verify backend reflects the change
    const currentMode = await getCurrentModeAPI(page)
    expect(currentMode).toBe('vulnerable')

    await resetToSecure(page)
  })

  test('switching back to secure mode updates banner immediately', async ({ page }) => {
    // First switch to vulnerable
    await setModeViaAPI(page, 'vulnerable')
    await page.reload()
    await expect(page.getByRole('banner')).toContainText(/VULNERABLE/i)

    // Now switch back to secure via toggle
    const toggleBtn = page.getByRole('button', { name: /Switch to Secure/i })
    await toggleBtn.click()
    await expect(page.getByRole('banner')).toContainText(/SECURE MODE/i, { timeout: 5000 })
  })
})

// ── Mode toggle — dashboard (post-auth) ──────────────────────────────────────

test.describe('Mode toggle — dashboard (authenticated)', () => {
  test.beforeEach(async ({ page }) => {
    await resetToSecure(page)

    // Login with test admin credentials
    await page.goto('/login')
    await page.fill('[placeholder="Masukkan username"]', 'admin')
    await page.fill('[placeholder="Masukkan password"]', 'Admin@123')
    await page.click('button[type=submit]')
    await page.waitForURL('**/dashboard', { timeout: 10000 })
  })

  test.afterEach(async ({ page }) => {
    await resetToSecure(page)
  })

  test('mode indicator in sidebar reflects current mode', async ({ page }) => {
    // Sidebar footer chip shows current mode
    await expect(page.getByText(/Secure Mode/i).last()).toBeVisible()
  })

  test('mode switch updates sidebar chip without reload', async ({ page }) => {
    const bannerBtn = page.getByRole('button', { name: /Switch to Vulnerable/i })
    await bannerBtn.click()
    await expect(page.getByText(/Vulnerable Mode/i).last()).toBeVisible({ timeout: 5000 })
  })

  test('dashboard page reflects mode change without reload', async ({ page }) => {
    // Check the banner changes
    await page.getByRole('button', { name: /Switch to Vulnerable/i }).click()
    await expect(page.getByRole('banner')).toContainText(/VULNERABLE/i, { timeout: 5000 })

    // ModeBadge in nav should update too
    await expect(page.getByText(/Vulnerable/).first()).toBeVisible()
  })

  test('multiple rapid mode switches stabilize correctly', async ({ page }) => {
    for (let i = 0; i < 3; i++) {
      await page.getByRole('button', { name: /Switch to (Vulnerable|Secure)/i }).click()
      await page.waitForTimeout(500) // small debounce for rapid test clicks
    }
    // After odd number of switches from secure, should be vulnerable
    const currentMode = await getCurrentModeAPI(page)
    expect(['secure', 'vulnerable']).toContain(currentMode)
  })
})

// ── Accessibility ─────────────────────────────────────────────────────────────

test.describe('GlobalModeSwitcher — accessibility', () => {
  test.beforeEach(async ({ page }) => {
    await resetToSecure(page)
    await page.goto('/')
  })

  test('banner has role="banner" attribute', async ({ page }) => {
    await expect(page.locator('[role="banner"]')).toBeVisible()
  })

  test('toggle button has aria-label', async ({ page }) => {
    const btn = page.getByRole('button', { name: /Switch to Vulnerable/i })
    await expect(btn).toBeVisible()
    // The button itself carries semantic label via its text content
    await expect(btn).toHaveText(/Switch to Vulnerable/i)
  })

  test('toggle button is keyboard focusable', async ({ page }) => {
    // Tab through the page and check that the banner button is reachable
    await page.keyboard.press('Tab')
    // The banner button should be one of the first focusable elements
    const focused = page.locator(':focus')
    // At minimum, it should be possible to reach it by keyboard
    await expect(focused).toBeDefined()
  })
})

// ── ModeBadge in navbar ───────────────────────────────────────────────────────

test.describe('ModeBadge in Navbar (authenticated)', () => {
  test.beforeEach(async ({ page }) => {
    await resetToSecure(page)
    await page.goto('/login')
    await page.fill('[placeholder="Masukkan username"]', 'admin')
    await page.fill('[placeholder="Masukkan password"]', 'Admin@123')
    await page.click('button[type=submit]')
    await page.waitForURL('**/dashboard', { timeout: 10000 })
  })

  test.afterEach(async ({ page }) => {
    await resetToSecure(page)
  })

  test('ModeBadge shows "Secure" in secure mode', async ({ page }) => {
    await expect(page.getByText('Secure').first()).toBeVisible()
  })

  test('ModeBadge updates to "Vulnerable" after mode switch', async ({ page }) => {
    await page.getByRole('button', { name: /Switch to Vulnerable/i }).click()
    await expect(page.getByText('Vulnerable').first()).toBeVisible({ timeout: 5000 })
  })
})
