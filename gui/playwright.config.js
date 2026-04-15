import { defineConfig } from '@playwright/test'
import { resolve } from 'path'

const ROOT = resolve(import.meta.dirname, '..')

export default defineConfig({
  testDir: './e2e',
  timeout: 30000,
  use: {
    baseURL: 'http://127.0.0.1:8374',
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure',
  },
  projects: [
    { name: 'chromium', use: { browserName: 'chromium' } },
  ],
  webServer: {
    command: `${ROOT}/glitch --workspace ${ROOT}/test-workspace workflow gui`,
    url: 'http://127.0.0.1:8374',
    reuseExistingServer: true,
    timeout: 15000,
  },
})
