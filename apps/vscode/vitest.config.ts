import { defineConfig } from 'vitest/config'

// The integration specs need a VS Code host (npm run test:integration); vitest runs the unit specs of src.
export default defineConfig({
  test: { include: ['src/**/*.spec.ts'] },
})
