import * as vscode from 'vscode'
import { destinationFor, LEAVE, MOVE_QUESTION, moveChoices, type MoveChoice } from './move-choices.policy.ts'
import type { RootProject } from './repository-offer.contract.ts'
import { say } from '../words/index.ts'

const TITLE = say('vscode.adopt-offer.monosmith-bring-in')

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
    { title: `${TITLE} (1/3): ${MOVE_QUESTION}`, placeHolder: say('vscode.adopt-offer.is-at-the-root-smith-manages', { manifest: project.manifest }) },
  )
  if (picked === undefined) return undefined
  const dir = picked.choice.key === 'none' ? LEAVE : await askName(picked.choice, project)
  if (dir === undefined) return undefined

  const ticked = await vscode.window.showQuickPick(
    [{ label: say('dialog.commit.label'), description: say('vscode.adopt-offer.unticked-smith-commits-nothing-unless-you'), picked: false }],
    { title: `${TITLE} (3/3): commit?`, canPickMany: true, placeHolder: say('vscode.adopt-offer.enter-with-nothing-ticked-leaves-the') },
  )

  return ticked === undefined ? undefined : { dir, commit: ticked.length > 0 }
}

async function askName (choice: MoveChoice, project: RootProject): Promise<string | undefined> {
  const whole = choice.key === 'other'

  const typed = await vscode.window.showInputBox({
    title:         `${TITLE} (2/3): ${say(whole ? 'vscode.adopt-offer.folder' : 'vscode.adopt-offer.name')}`,
    prompt:        whole ? say('vscode.adopt-offer.the-folder-from-the-repository-root') : say('vscode.adopt-offer.the-name-of-the-folder-in', { key: choice.key }),
    ...(whole ? { placeHolder: say('vscode.adopt-offer.for-example-tools', { name: project.name }) } : { value: project.name }),
    validateInput: (value) => {
      const result = destinationFor(choice, value)

      return 'problem' in result ? result.problem : undefined
    },
  })
  if (typed === undefined) return undefined
  const result = destinationFor(choice, typed)

  return 'dir' in result ? result.dir : undefined
}
