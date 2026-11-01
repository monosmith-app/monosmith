import type { DeriveAction, DeriveStatus } from './derive-report.contract.ts'

/**
 * The arguments of `smith derive` for an action. The file goes in `--flag=file`
 * form, so a name that starts with a dash is never read as an option.
 * @param action - what to do
 * @param file - the file, relative to the folder; needed by every action but `check` and a plain `apply`
 * @returns the arguments; `apply` and the three choices pass `--yes`, because the person agreed in the editor
 */
export function deriveArguments (action: DeriveAction, file = ''): string[] {
  switch (action) {
    case 'check': {
      return ['derive', '--check']
    }
    case 'apply': {
      return ['derive', '--yes']
    }
    case 'explain': {
      return ['derive', `--explain=${file}`]
    }
    default: {
      return ['derive', `--${action}=${file}`, '--yes']
    }
  }
}

const HELD_BACK: ReadonlySet<DeriveStatus> = new Set(['drift', 'diverged', 'broken'])

/**
 * Whether a plain `smith derive` holds the file back until the person chooses.
 * @param status - the state of the file
 */
export function isHeldBack (status: DeriveStatus): boolean {
  return HELD_BACK.has(status)
}

/**
 * What Apply means for a file. A stale, missing or unrecorded file is written
 * by a plain `smith derive`. A drifted, diverged or broken file is held back
 * by it, so Apply overwrites that one file with `--replace`.
 * @param status - the state of the file
 */
export function applyActionFor (status: DeriveStatus): 'apply' | 'replace' {
  return isHeldBack(status) ? 'replace' : 'apply'
}
