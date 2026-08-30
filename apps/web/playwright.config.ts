import { defineConfig, devices } from "@playwright/test";

const URL_BASE = process.env.URL_E2E ?? "http://localhost:3000";

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: 1,
  reporter: process.env.CI ? [["html", { open: "never" }], ["list"]] : "list",

  use: {
    baseURL: URL_BASE,
    trace: "on-first-retry",
  },

  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],

  // Playwright levanta Next, no la API de Go. Mezclar dos runtimes en un
  // webServer es donde estos setups se vuelven frágiles: la API se levanta
  // afuera, con `make run` en local y con un paso propio en el CI.
  webServer: {
    command: "pnpm dev",
    url: URL_BASE,
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
});
