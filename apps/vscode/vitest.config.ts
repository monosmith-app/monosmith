import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vitest/config'

// The integration specs need a VS Code host (npm run test:integration); vitest runs the unit specs of src.
// `vscode` resolves to an empty stand-in: no unit spec calls the editor.
export default defineConfig({
  resolve: { alias: { vscode: fileURLToPath(new URL('vitest.vscode-stub.ts', import.meta.url)) } },
  test:    { include: ['src/**/*.spec.ts'] },
})
