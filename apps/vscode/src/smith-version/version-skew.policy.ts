import type { SmithSource } from '../smith-client/index.ts'
import { sourceName } from './smith-origin.mapper.ts'

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

  return check.terminalVersion === check.editorVersion ? undefined : `MonoSmith runs smith ${check.editorVersion} from ${sourceName(check.editorSource)}, but a terminal here runs smith ${check.terminalVersion} from the PATH. Results can differ.`
}
