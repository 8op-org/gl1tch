import { defineConfig } from '@playwright/test'
import { resolve } from 'path'
import { homedir } from 'os'

const ROOT = resolve(import.meta.dirname, '..')
const WORKSPACE = process.env.GLITCH_TEST_WORKSPACE || resolve(homedir(), 'Projects/stokagent')

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
    command: `${ROOT}/glitch --workspace ${WORKSPACE} workflow gui`,
    url: 'http://127.0.0.1:8374',
    reuseExistingServer: true,
    timeout: 15000,
  },
})
