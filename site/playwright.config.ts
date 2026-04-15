import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  timeout: 15000,
  use: {
    baseURL: 'http://localhost:4322',
  },
  webServer: {
    command: 'npx astro dev --port 4322',
    port: 4322,
    reuseExistingServer: true,
  },
});
