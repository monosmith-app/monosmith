// Lint AND formatting for the workspace: @monosmith/eslint-config (built from
// packages/eslint-config) carries the TypeScript rules, JavaScript Standard
// Style and the vertical-slice rules. Add your own blocks after the spread;
// later blocks win.
import monosmith from '@monosmith/eslint-config'

export default [
  ...monosmith(),
  // The jest-preset tests create and delete throwaway projects here while the lint task runs.
  { ignores: ['**/.fixture-*/**'] },
  // A typical single-project repository that the adopt tests split: its code is not ours to lint.
  { ignores: ['apps/smith/adoption/testdata/**'] },
  // Written by `go run ./wordgen` in libs/words: the generator owns its format.
  { ignores: ['apps/vscode/src/words/smith-words.config.ts'] },
]
