import type { RootProject } from './repository-offer.contract.ts'
import { say } from '../words/index.ts'

/** The `--move` value that keeps the project where it is. */
export const LEAVE = 'none'

/** The question smith asks, word for word (`smith adopt --init` and the wizard use it too). */
export const MOVE_QUESTION = say('vscode.adopt-offer.move-it-into-a-folder-of')

/** One answer of the question. */
export interface MoveChoice {
  key:       'apps' | 'packages' | 'libs' | 'other' | 'none'
  /** The option as the pick shows it. */
  label:     string
  /** The folder, undefined for `other` (the person types it) and `none`. */
  dir?:      string
  suggested: boolean
}

const NAME_PATTERN = /^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$/
const FOLDER_SEGMENT = /^[A-Z0-9][\w.-]*$/i

/**
 * The five answers, in the order smith numbers them. The suggested one is an
 * app folder when the project looks like a program, else a package folder.
 * @param project - the project found at the root
 */
export function moveChoices (project: RootProject): MoveChoice[] {
  const suggested = project.app ? 'apps' : 'packages'

  return [
    ...(['apps', 'packages', 'libs'] as const).map((key) => ({ key, label: `${key}/${project.name}`, dir: `${key}/${project.name}`, suggested: key === suggested })),
    { key: 'other', label: say('vscode.adopt-offer.another-folder'), suggested: false },
    { key: 'none', label: say('vscode.adopt-offer.no-leave-it'), suggested: false },
  ]
}

/**
 * Tells why a name cannot be a project name.
 * @param name - what the person typed
 * @returns a sentence, or undefined when the name is fine
 */
export function checkName (name: string): string | undefined {
  return name.length <= 50 && NAME_PATTERN.test(name) ? undefined : say('vscode.adopt-offer.must-be-lowercase-letters-and-digits', { stringify: JSON.stringify(name) })
}

/**
 * Checks a folder typed relative to the repository root.
 * @param folder - what the person typed
 * @returns the cleaned folder, or a sentence saying why it cannot be used
 */
export function checkFolder (folder: string): { dir: string } | { problem: string } {
  const raw = folder.trim().replaceAll('\\', '/')
  const bad = (why: string): { problem: string } => ({ problem: say('vscode.adopt-offer.is-not-a-folder-smith-can', { stringify: JSON.stringify(folder), why }) })
  if (['', '.', './'].includes(raw)) return bad(say('vscode.adopt-offer.it-is-the-repository-root-where'))
  if (raw.startsWith('/') || /^[A-Z]:/i.test(raw)) return bad(say('vscode.adopt-offer.it-must-be-relative-to-the'))
  const segments = raw.split('/').filter((segment) => segment !== '' && segment !== '.')
  for (const segment of segments) {
    if (segment === '..') return bad(say('vscode.adopt-offer.it-leaves-the-repository'))
    if (segment.startsWith('.')) return bad(say('vscode.adopt-offer.a-folder-that-starts-with-a'))
    if (!FOLDER_SEGMENT.test(segment)) return bad(say('vscode.adopt-offer.use-letters-digits-dot-hyphen-and'))
  }

  return { dir: segments.join('/') }
}

/**
 * The folder for a choice.
 * @param choice - the answer
 * @param typed - the name (apps, packages, libs) or the whole folder (other)
 * @returns the folder, or a sentence saying why it cannot be used
 */
export function destinationFor (choice: MoveChoice, typed: string): { dir: string } | { problem: string } {
  if (choice.key === 'other') return checkFolder(typed)
  const problem = checkName(typed)

  return problem === undefined ? { dir: `${choice.key}/${typed}` } : { problem }
}
