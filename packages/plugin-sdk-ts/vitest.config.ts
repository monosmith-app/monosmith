import { defineConfig } from 'vitest/config'

// Unit specs sit beside what they test (src, tools); the integration specs
// start real processes (the Go plugins, smith) and get a longer timeout.
export default defineConfig({
  test: {
    include:     ['src/**/*.spec.ts', 'tools/**/*.spec.ts', 'integration/**/*.spec.ts'],
    testTimeout: 120_000,
    hookTimeout: 120_000,
  },
})
