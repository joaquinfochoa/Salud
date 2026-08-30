import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    // e2e/ es de Playwright. Sin esta exclusión vitest levanta reservar.spec.ts
    // y falla con "Playwright Test did not expect test() to be called here":
    // el patrón por defecto de vitest incluye *.spec.ts, así que los dos
    // corredores se pelean por el mismo archivo.
    exclude: ["node_modules/**", ".next/**", "e2e/**"],
  },
});
