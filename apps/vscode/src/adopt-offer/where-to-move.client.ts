import * as vscode from 'vscode'
import { destinationFor, LEAVE, MOVE_QUESTION, moveChoices, type MoveChoice } from './move-choices.policy.ts'
import type { RootProject } from './repository-offer.contract.ts'

const TITLE = 'MonoSmith: bring in'

/** What the person decided. */
export interface MoveAnswer {
  /** The folder the project goes into, or `none` to leave it at the root. */
  dir:    string
  /** True when they ticked "Commit this change so it can be undone". */
  commit: boolean
}

/**
 * Asks where the project at the repository root goes, the question
 * `smith adopt --init` asks in a terminal: a quick pick of five answers with the
 * suggested one marked, an input box for the name (or the whole folder) and an
 * unticked "Commit this change so it can be undone" box. Nothing is changed.
 * @param project - the project found at the root
 * @returns the answer, or undefined when the person pressed Escape at any step
 */
export async function askWhereToMove (project: RootProject): Promise<MoveAnswer | undefined> {
  const picked = await vscode.window.showQuickPick(
    moveChoices(project).map((choice) => ({ label: choice.label, description: choice.suggested ? 'suggested' : '', choice })),
    { title: `${TITLE} (1/3): ${MOVE_QUESTION}`, placeHolder: `${project.manifest} is at the root; smith manages one folder per project and does not pick it for you` },
  )
  if (picked === undefined) return undefined
  const dir = picked.choice.key === 'none' ? LEAVE : await askName(picked.choice, project)
  if (dir === undefined) return undefined

  const ticked = await vscode.window.showQuickPick(
    [{ label: 'Commit this change so it can be undone', description: 'unticked: smith commits nothing unless you tick this', picked: false }],
    { title: `${TITLE} (3/3): commit?`, canPickMany: true, placeHolder: 'Enter with nothing ticked leaves the change uncommitted' },
  )

  return ticked === undefined ? undefined : { dir, commit: ticked.length > 0 }
}

async function askName (choice: MoveChoice, project: RootProject): Promise<string | undefined> {
  const whole = choice.key === 'other'

  const typed = await vscode.window.showInputBox({
    title:         `${TITLE} (2/3): ${whole ? 'folder' : 'name'}`,
    prompt:        whole ? 'The folder, from the repository root' : `The name of the folder in ${choice.key}/`,
    ...(whole ? { placeHolder: `for example tools/${project.name}` } : { value: project.name }),
    validateInput: (value) => {
      const result = destinationFor(choice, value)

      return 'problem' in result ? result.problem : undefined
    },
  })
  if (typed === undefined) return undefined
  const result = destinationFor(choice, typed)

  return 'dir' in result ? result.dir : undefined
}
