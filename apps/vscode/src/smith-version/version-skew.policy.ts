import type { SmithSource } from '../smith-client/index.ts'
import { sourceName } from './smith-origin.mapper.ts'
import { say } from '../words/index.ts'

/** The smith the editor runs and the one a terminal would run. */
export interface SkewCheck {
  editorSmith:     string
  editorVersion:   string
  editorSource:    SmithSource
  terminalSmith:   string | undefined
  terminalVersion: string | undefined
}

/**
 * Whether the editor runs a different smith than the terminal does. That is
 * only worth saying when they are two executables of two versions: the same
 * file, or an unknown terminal version, is not a difference.
 * @returns the warning, or undefined when there is nothing to say
 */
export function judgeSkew (check: SkewCheck): string | undefined {
  if (check.terminalSmith === undefined || check.terminalVersion === undefined || check.editorSource === 'path' || check.terminalSmith === check.editorSmith) return undefined

  return check.terminalVersion === check.editorVersion ? undefined : say('vscode.smith-version.monosmith-runs-smith-from-but-a', { editorVersion: check.editorVersion, sourceName: sourceName(check.editorSource), terminalVersion: check.terminalVersion })
}
