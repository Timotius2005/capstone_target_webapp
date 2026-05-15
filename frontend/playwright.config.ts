import { defineConfig, devices } from '@playwright/test'

// In CI, add a JUnit reporter so GitHub Actions can parse test results.
// Locally, keep the concise list reporter instead.
const reporters: Parameters<typeof defineConfig>[0]['reporter'] = process.env.CI
  ? [
      ['junit', { outputFile: '../reports/playwright-junit.xml' }],
      ['html', { outputFolder: '../reports/playwright-report', open: 'never' }],
      ['json', { outputFile: '../reports/playwright-results.json' }],
      ['list'],
    ]
  : [
      ['html', { outputFolder: '../reports/playwright-report', open: 'never' }],
      ['json', { outputFile: '../reports/playwright-results.json' }],
      ['list'],
    ]

export default defineConfig({
  testDir: './e2e',
  globalSetup: require.resolve('./e2e/global-setup'),
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,

  reporter: reporters,

  use: {
    baseURL: process.env.PLAYWRIGHT_BASE_URL ?? 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },

  projects: [
    { name: 'chromium', use: { ...devices['Desktop Chrome'] } },
  ],
})
