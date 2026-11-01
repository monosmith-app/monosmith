import { describeFailure, type SmithRun } from '../smith-run/index.ts'
import type { InstallOutcome, InstallPreview } from './install-run.contract.ts'

const NO_TERMINAL = 'there is no terminal to ask on'

/** The `● Add .mcp.json` lines of smith's output: what each file would get. */
function actionLines (stdout: string): string[] {
  return stdout.split(/\r?\n/).filter((line) => line.startsWith('● ')).map((line) => line.slice(2).trim())
}

function refusal (run: SmithRun): InstallOutcome {
  return { kind: 'refused', message: describeFailure(run) }
}

/**
 * Reads a run without `--yes`: smith prints the diffs and stops because no
 * terminal can answer. Anything else it stopped for (an edited file, a file
 * with comments) is a refusal, and its words are kept.
 * @returns the preview, `current` when there is nothing to write, or the refusal
 */
export function readPreview (run: SmithRun): InstallPreview | InstallOutcome {
  if (run.code === 0 && run.stdout.includes('already installed')) return { kind: 'current' }
  if (run.code === 1 && run.stderr.includes(NO_TERMINAL)) {
    const actions = actionLines(run.stdout)
    if (actions.length > 0) return { actions, diff: run.stdout.trim() }
  }

  return refusal(run)
}

/** Reads a run with `--yes`. */
export function readWrite (run: SmithRun): InstallOutcome {
  if (run.code !== 0) return refusal(run)
  const lines = run.stdout.split(/\r?\n/).filter((line) => line.startsWith('✓ Wrote ')).map((line) => line.replace(/^✓\s*/, ''))

  return lines.length === 0 ? { kind: 'current' } : { kind: 'written', lines }
}

/** Tells a preview from the outcomes that need no asking. */
export function isPreview (value: InstallPreview | InstallOutcome): value is InstallPreview {
  return 'actions' in value
}
