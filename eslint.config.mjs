// Lint AND formatting for the workspace: @monosmith/eslint-config (built from
// packages/eslint-config) carries the TypeScript rules, JavaScript Standard
// Style and the vertical-slice rules. Add your own blocks after the spread;
// later blocks win.
import monosmith from '@monosmith/eslint-config'

export default [
  ...monosmith(),
  // The jest-preset tests create and delete throwaway projects here while the lint task runs.
  { ignores: ['**/.fixture-*/**'] },
]
